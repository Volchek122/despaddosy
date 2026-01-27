package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"despaddosy/pkg/config"
	"despaddosy/pkg/logging"
	"despaddosy/pkg/store"
	"despaddosy/pkg/tracing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

//go:embed ../../web/templates/*.html
var templateFS embed.FS

//go:embed ../../web/static/*
var staticFS embed.FS

type server struct {
	cfg       config.SocialConfig
	db        *pgxpool.Pool
	redis     *redis.Client
	templates *template.Template
}

type user struct {
	ID       int
	Email    string
	Username string
	Role     string
	Bio      string
	Avatar   string
	Banned   bool
}

type post struct {
	ID        int
	UserID    int
	Username  string
	Content   string
	ImageURL  string
	CreatedAt time.Time
	Likes     int
	Comments  int
}

type comment struct {
	ID        int
	PostID    int
	UserID    int
	Username  string
	Content   string
	CreatedAt time.Time
}

type report struct {
	ID        int
	PostID    int
	Reason    string
	Status    string
	Note      string
	CreatedAt time.Time
}

type authClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func main() {
	cfgPath := os.Getenv("SOCIALD_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/sociald.yaml"
	}
	cfg, err := config.LoadSocialConfig(cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load sociald config")
	}
	logging.Init(cfg.ServiceName)

	ctx := context.Background()
	shutdown, err := tracing.Init(ctx, cfg.ServiceName, cfg.Observability.OTLPEndpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init tracing")
	}
	defer func() { _ = shutdown(context.Background()) }()

	db, err := store.NewPostgres(ctx, cfg.Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
	}
	redisClient := store.NewRedis(cfg.Redis.Addr)

	tmpl, err := template.ParseFS(templateFS, "../../web/templates/*.html")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse templates")
	}

	srv := &server{cfg: cfg, db: db, redis: redisClient, templates: tmpl}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", srv.feed)
	mux.HandleFunc("/login", srv.login)
	mux.HandleFunc("/register", srv.register)
	mux.HandleFunc("/logout", srv.logout)
	mux.HandleFunc("/u/", srv.profile)
	mux.HandleFunc("/post/", srv.viewPost)
	mux.HandleFunc("/api/posts", srv.posts)
	mux.HandleFunc("/api/posts/", srv.postActions)
	mux.HandleFunc("/api/reports", srv.reports)
	mux.HandleFunc("/admin", srv.adminDashboard)
	mux.HandleFunc("/admin/users", srv.adminUsers)
	mux.HandleFunc("/admin/posts", srv.adminPosts)
	mux.HandleFunc("/admin/reports", srv.adminReports)
	mux.HandleFunc("/admin/security", srv.adminSecurity)

	if cfg.Observability.MetricsAddr != "" {
		go func() {
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(cfg.Observability.MetricsAddr, metricsMux); err != nil {
				log.Fatal().Err(err).Msg("metrics server failed")
			}
		}()
	}

	wrapped := logging.RequestID(logging.AccessLog(mux))
	srvHTTP := &http.Server{Addr: cfg.Address, Handler: wrapped}
	log.Info().Str("addr", cfg.Address).Msg("sociald listening")
	if cfg.TLS.Enabled {
		srvHTTP.TLSConfig = socialTLS(cfg)
		if err := srvHTTP.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
			log.Fatal().Err(err).Msg("sociald stopped")
		}
	}
	if err := srvHTTP.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("sociald stopped")
	}
}

func socialTLS(cfg config.SocialConfig) *tls.Config {
	caCert, err := os.ReadFile(cfg.TLS.CAFile)
	if err != nil {
		return &tls.Config{}
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		ClientCAs:  pool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
}

func (s *server) feed(w http.ResponseWriter, r *http.Request) {
	posts, _ := s.listPosts(r.Context())
	user := s.currentUser(r)
	data := map[string]any{"Posts": posts, "User": user, "CSRF": s.ensureCSRF(w, r)}
	_ = s.templates.ExecuteTemplate(w, "feed.html", data)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]any{"CSRF": s.ensureCSRF(w, r)}
		_ = s.templates.ExecuteTemplate(w, "login.html", data)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.verifyCSRF(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := s.findUserByEmail(r.Context(), email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Banned {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	token, _ := s.issueToken(user.ID, user.Role)
	s.setAuthCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := map[string]any{"CSRF": s.ensureCSRF(w, r)}
		_ = s.templates.ExecuteTemplate(w, "register.html", data)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !s.verifyCSRF(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	if email == "" || username == "" || len(password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := s.db.Exec(r.Context(), "INSERT INTO users(email, username, password_hash, role) VALUES ($1,$2,$3,'user')", email, username, string(hash))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	s.clearAuthCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) profile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/u/")
	usr, err := s.findUserByUsername(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	posts, _ := s.listPostsByUser(r.Context(), usr.ID)
	data := map[string]any{"Profile": usr, "Posts": posts, "User": s.currentUser(r)}
	_ = s.templates.ExecuteTemplate(w, "profile.html", data)
}

func (s *server) viewPost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/post/")
	id, _ := strconv.Atoi(idStr)
	post, err := s.getPost(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	comments, _ := s.listComments(r.Context(), id)
	data := map[string]any{"Post": post, "Comments": comments, "User": s.currentUser(r), "CSRF": s.ensureCSRF(w, r)}
	_ = s.templates.ExecuteTemplate(w, "post.html", data)
}

func (s *server) posts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		posts, _ := s.listPosts(r.Context())
		respondJSON(w, posts)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user := s.requireUser(w, r)
	if user == nil {
		return
	}
	if !s.verifyCSRF(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")
	image := r.FormValue("image_url")
	if content == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err := s.db.Exec(r.Context(), "INSERT INTO posts(user_id, content, image_url) VALUES ($1,$2,$3)", user.ID, content, image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) postActions(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/posts/"), "/")
	if len(pathParts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	postID, _ := strconv.Atoi(pathParts[0])
	action := pathParts[1]
	user := s.requireUser(w, r)
	if user == nil {
		return
	}
	switch action {
	case "like":
		if r.Method == http.MethodPost {
			if !s.verifyCSRF(r) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = s.db.Exec(r.Context(), "INSERT INTO likes(user_id, post_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", user.ID, postID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete {
			_, _ = s.db.Exec(r.Context(), "DELETE FROM likes WHERE user_id=$1 AND post_id=$2", user.ID, postID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	case "comments":
		if r.Method == http.MethodGet {
			comments, _ := s.listComments(r.Context(), postID)
			respondJSON(w, comments)
			return
		}
		if r.Method == http.MethodPost {
			if !s.verifyCSRF(r) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			content := r.FormValue("content")
			if content == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = s.db.Exec(r.Context(), "INSERT INTO comments(post_id, user_id, content) VALUES ($1,$2,$3)", postID, user.ID, content)
			http.Redirect(w, r, "/post/"+strconv.Itoa(postID), http.StatusSeeOther)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *server) reports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user := s.requireUser(w, r)
	if user == nil {
		return
	}
	if !s.verifyCSRF(r) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	postID, _ := strconv.Atoi(r.FormValue("post_id"))
	reason := r.FormValue("reason")
	_, _ = s.db.Exec(r.Context(), "INSERT INTO reports(reporter_id, post_id, reason, status) VALUES ($1,$2,$3,'open')", user.ID, postID, reason)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	if s.requireAdmin(w, r) == nil {
		return
	}
	data := map[string]any{"User": s.currentUser(r)}
	_ = s.templates.ExecuteTemplate(w, "admin.html", data)
}

func (s *server) adminUsers(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("user_id"))
		action := r.FormValue("action")
		if action == "ban" {
			_, _ = s.db.Exec(r.Context(), "UPDATE users SET banned=true WHERE id=$1", id)
		}
		if action == "unban" {
			_, _ = s.db.Exec(r.Context(), "UPDATE users SET banned=false WHERE id=$1", id)
		}
		if action == "promote" {
			_, _ = s.db.Exec(r.Context(), "UPDATE users SET role='admin' WHERE id=$1", id)
		}
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}
	rows, _ := s.db.Query(r.Context(), "SELECT id, email, username, role, banned FROM users ORDER BY id")
	users := []user{}
	for rows.Next() {
		var u user
		_ = rows.Scan(&u.ID, &u.Email, &u.Username, &u.Role, &u.Banned)
		users = append(users, u)
	}
	data := map[string]any{"Users": users, "User": admin, "CSRF": s.ensureCSRF(w, r)}
	_ = s.templates.ExecuteTemplate(w, "admin_users.html", data)
}

func (s *server) adminPosts(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("post_id"))
		action := r.FormValue("action")
		if action == "hide" {
			_, _ = s.db.Exec(r.Context(), "UPDATE posts SET hidden=true WHERE id=$1", id)
		}
		if action == "unhide" {
			_, _ = s.db.Exec(r.Context(), "UPDATE posts SET hidden=false WHERE id=$1", id)
		}
		http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
		return
	}
	posts, _ := s.listPosts(r.Context())
	data := map[string]any{"Posts": posts, "User": admin, "CSRF": s.ensureCSRF(w, r)}
	_ = s.templates.ExecuteTemplate(w, "admin_posts.html", data)
}

func (s *server) adminReports(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("report_id"))
		action := r.FormValue("action")
		note := r.FormValue("note")
		if action == "close" {
			_, _ = s.db.Exec(r.Context(), "UPDATE reports SET status='closed', admin_note=$1 WHERE id=$2", note, id)
		}
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
		return
	}
	rows, _ := s.db.Query(r.Context(), "SELECT id, post_id, reason, status, admin_note, created_at FROM reports ORDER BY created_at DESC")
	reports := []report{}
	for rows.Next() {
		var rep report
		_ = rows.Scan(&rep.ID, &rep.PostID, &rep.Reason, &rep.Status, &rep.Note, &rep.CreatedAt)
		reports = append(reports, rep)
	}
	data := map[string]any{"Reports": reports, "User": admin, "CSRF": s.ensureCSRF(w, r)}
	_ = s.templates.ExecuteTemplate(w, "admin_reports.html", data)
}

func (s *server) adminSecurity(w http.ResponseWriter, r *http.Request) {
	admin := s.requireAdmin(w, r)
	if admin == nil {
		return
	}
	events := []map[string]any{}
	results, _ := s.redis.LRange(r.Context(), "secevents", 0, 50).Result()
	for _, item := range results {
		var data map[string]any
		if err := json.Unmarshal([]byte(item), &data); err == nil {
			events = append(events, data)
		}
	}
	data := map[string]any{"Events": events, "User": admin}
	_ = s.templates.ExecuteTemplate(w, "admin_security.html", data)
}

func (s *server) listPosts(ctx context.Context) ([]post, error) {
	rows, err := s.db.Query(ctx, `SELECT p.id, p.user_id, u.username, p.content, p.image_url, p.created_at,
		(SELECT COUNT(1) FROM likes l WHERE l.post_id=p.id),
		(SELECT COUNT(1) FROM comments c WHERE c.post_id=p.id)
		FROM posts p JOIN users u ON p.user_id=u.id WHERE p.hidden=false ORDER BY p.created_at DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := []post{}
	for rows.Next() {
		var p post
		_ = rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Content, &p.ImageURL, &p.CreatedAt, &p.Likes, &p.Comments)
		posts = append(posts, p)
	}
	return posts, nil
}

func (s *server) listPostsByUser(ctx context.Context, userID int) ([]post, error) {
	rows, err := s.db.Query(ctx, `SELECT p.id, p.user_id, u.username, p.content, p.image_url, p.created_at,
		(SELECT COUNT(1) FROM likes l WHERE l.post_id=p.id),
		(SELECT COUNT(1) FROM comments c WHERE c.post_id=p.id)
		FROM posts p JOIN users u ON p.user_id=u.id WHERE p.user_id=$1 ORDER BY p.created_at DESC LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := []post{}
	for rows.Next() {
		var p post
		_ = rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Content, &p.ImageURL, &p.CreatedAt, &p.Likes, &p.Comments)
		posts = append(posts, p)
	}
	return posts, nil
}

func (s *server) getPost(ctx context.Context, id int) (post, error) {
	var p post
	err := s.db.QueryRow(ctx, `SELECT p.id, p.user_id, u.username, p.content, p.image_url, p.created_at,
		(SELECT COUNT(1) FROM likes l WHERE l.post_id=p.id),
		(SELECT COUNT(1) FROM comments c WHERE c.post_id=p.id)
		FROM posts p JOIN users u ON p.user_id=u.id WHERE p.id=$1`, id).Scan(&p.ID, &p.UserID, &p.Username, &p.Content, &p.ImageURL, &p.CreatedAt, &p.Likes, &p.Comments)
	return p, err
}

func (s *server) listComments(ctx context.Context, postID int) ([]comment, error) {
	rows, err := s.db.Query(ctx, `SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c JOIN users u ON c.user_id=u.id WHERE c.post_id=$1 ORDER BY c.created_at`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := []comment{}
	for rows.Next() {
		var c comment
		_ = rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.Content, &c.CreatedAt)
		comments = append(comments, c)
	}
	return comments, nil
}

func (s *server) findUserByEmail(ctx context.Context, email string) (*userWithHash, error) {
	var u userWithHash
	err := s.db.QueryRow(ctx, "SELECT id, email, username, password_hash, role, bio, avatar_url, banned FROM users WHERE email=$1", email).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Role, &u.Bio, &u.Avatar, &u.Banned)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *server) findUserByUsername(ctx context.Context, username string) (*user, error) {
	var u user
	err := s.db.QueryRow(ctx, "SELECT id, email, username, role, bio, avatar_url, banned FROM users WHERE username=$1", username).
		Scan(&u.ID, &u.Email, &u.Username, &u.Role, &u.Bio, &u.Avatar, &u.Banned)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *server) currentUser(r *http.Request) *user {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return nil
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &authClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil
	}
	claims := token.Claims.(*authClaims)
	id, _ := strconv.Atoi(claims.UserID)
	usr, err := s.findUserByID(r.Context(), id)
	if err != nil {
		return nil
	}
	return usr
}

func (s *server) requireUser(w http.ResponseWriter, r *http.Request) *user {
	user := s.currentUser(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	if user.Banned {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}
	return user
}

func (s *server) requireAdmin(w http.ResponseWriter, r *http.Request) *user {
	user := s.requireUser(w, r)
	if user == nil {
		return nil
	}
	if user.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}
	return user
}

func (s *server) findUserByID(ctx context.Context, id int) (*user, error) {
	var u user
	err := s.db.QueryRow(ctx, "SELECT id, email, username, role, bio, avatar_url, banned FROM users WHERE id=$1", id).
		Scan(&u.ID, &u.Email, &u.Username, &u.Role, &u.Bio, &u.Avatar, &u.Banned)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *server) issueToken(userID int, role string) (string, error) {
	claims := authClaims{
		UserID: strconv.Itoa(userID),
		Roles:  []string{role},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *server) setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func (s *server) clearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1}
	http.SetCookie(w, cookie)
}

func respondJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *server) ensureCSRF(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("csrf_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	value := strconv.FormatInt(time.Now().UnixNano(), 36)
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    value,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	return value
}

func (s *server) verifyCSRF(r *http.Request) bool {
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return false
	}
	return r.FormValue("csrf_token") == cookie.Value
}

type userWithHash struct {
	user
	PasswordHash string
}

func (s *server) parseJSON(r *http.Request, target any) error {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("invalid content type")
	}
	return json.NewDecoder(r.Body).Decode(target)
}
