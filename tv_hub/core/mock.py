from typing import List, Optional
from .provider import ContentProvider
from .models import ContentItem, Season, Episode

class MockProvider(ContentProvider):
    def search(self, query: str) -> List[ContentItem]:
        # Return some dummy data
        return [
            ContentItem(
                title="Stranger Things",
                url="mock_url_1",
                poster_url="https://image.tmdb.org/t/p/w500/x2LSRK2Cm7MZhjluni1msUQ36HG.jpg",
                year="2016",
                type="series",
                rating="8.7",
                description="When a young boy vanishes, a small town uncovers a mystery involving secret experiments, terrifying supernatural forces, and one strange little girl."
            ),
            ContentItem(
                title="Breaking Bad",
                url="mock_url_2",
                poster_url="https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg",
                year="2008",
                type="series",
                rating="9.5",
                description="A high school chemistry teacher diagnosed with inoperable lung cancer turns to manufacturing and selling methamphetamine in order to secure his family's future."
            ),
            ContentItem(
                title="Inception",
                url="mock_url_3",
                poster_url="https://image.tmdb.org/t/p/w500/9gk7admal4zl248sKdtJn2od26.jpg",
                year="2010",
                type="movie",
                rating="8.8",
                description="A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O."
            )
        ]

    def get_details(self, url: str) -> ContentItem:
        # Return a detailed object based on the mock url
        if url == "mock_url_1":
            item = ContentItem(
                title="Stranger Things",
                url="mock_url_1",
                poster_url="https://image.tmdb.org/t/p/w500/x2LSRK2Cm7MZhjluni1msUQ36HG.jpg",
                year="2016",
                type="series",
                rating="8.7",
                description="When a young boy vanishes, a small town uncovers a mystery involving secret experiments, terrifying supernatural forces, and one strange little girl."
            )
            item.seasons = [
                Season(id="1", title="Season 1", episodes=[
                    Episode(title="Chapter One: The Vanishing of Will Byers", url="stream_s1e1", season_id="1", episode_id="1"),
                    Episode(title="Chapter Two: The Weirdo on Maple Street", url="stream_s1e2", season_id="1", episode_id="2")
                ]),
                Season(id="2", title="Season 2", episodes=[
                    Episode(title="Chapter One: MADMAX", url="stream_s2e1", season_id="2", episode_id="1")
                ])
            ]
            return item
        elif url == "mock_url_3":
             return ContentItem(
                title="Inception",
                url="mock_url_3",
                poster_url="https://image.tmdb.org/t/p/w500/9gk7admal4zl248sKdtJn2od26.jpg",
                year="2010",
                type="movie",
                rating="8.8",
                description="A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O."
            )
        return self.search("")[0]

    def get_stream(self, episode_url: str) -> str:
        # Return a test video URL (Big Buck Bunny or similar)
        return "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
