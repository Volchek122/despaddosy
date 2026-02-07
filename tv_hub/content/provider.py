from abc import ABC, abstractmethod

class ContentProvider(ABC):
    @abstractmethod
    def search(self, query):
        """Search for content. Returns a list of dictionaries."""
        pass

    @abstractmethod
    def get_popular(self):
        """Get popular/new content. Returns a list of dictionaries."""
        pass

    @abstractmethod
    def get_details(self, content_id):
        """Get details for a specific content item."""
        pass

    @abstractmethod
    def get_stream_url(self, content_id, season=None, episode=None):
        """Get the stream URL for playback."""
        pass

class MockProvider(ContentProvider):
    def __init__(self):
        self.movies = [
            {
                "id": "1",
                "title": "Cyberpunk: Edgerunners",
                "year": "2022",
                "poster_url": "https://static.hdrezka.ac/i/2022/9/13/f5d5226521362ri66q48o.jpg",
                "type": "series",
                "description": "In a dystopia riddled with corruption and cybernetic implants, a talented but reckless street kid strives to become a mercenary outlaw — an edgerunner.",
                "seasons": [1],
                "episodes": {1: 10}
            },
            {
                "id": "2",
                "title": "The Matrix",
                "year": "1999",
                "poster_url": "https://static.hdrezka.ac/i/2013/11/2/u306565176985br24p76o.jpg",
                "type": "movie",
                "description": "Thomas Anderson, a computer programmer, is led to fight an underground war against powerful computers who have constructed his entire reality with a system called the Matrix.",
                "seasons": [],
                "episodes": {}
            },
            {
                "id": "3",
                "title": "Breaking Bad",
                "year": "2008",
                "poster_url": "https://static.hdrezka.ac/i/2013/12/21/u05d315185934nc67q66o.jpg",
                "type": "series",
                "description": "A high school chemistry teacher diagnosed with inoperable lung cancer turns to manufacturing and selling methamphetamine in order to secure his family's future.",
                "seasons": [1, 2, 3, 4, 5],
                "episodes": {1: 7, 2: 13, 3: 13, 4: 13, 5: 16}
            },
             {
                "id": "4",
                "title": "Interstellar",
                "year": "2014",
                "poster_url": "https://static.hdrezka.ac/i/2014/11/10/u325515234556yb86p34o.jpg",
                "type": "movie",
                "description": "A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival.",
                "seasons": [],
                "episodes": {}
            }
        ]

    def search(self, query):
        query = query.lower()
        return [m for m in self.movies if query in m['title'].lower()]

    def get_popular(self):
        return self.movies

    def get_details(self, content_id):
        for m in self.movies:
            if m['id'] == content_id:
                return m
        return None

    def get_stream_url(self, content_id, season=None, episode=None):
        # Return a dummy video URL (Big Buck Bunny is standard for testing)
        return "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
