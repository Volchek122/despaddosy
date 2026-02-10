from dataclasses import dataclass
from typing import List, Optional

@dataclass
class Episode:
    title: str
    url: str  # URL to fetch the stream
    season_id: str
    episode_id: str

@dataclass
class Season:
    id: str
    title: str
    episodes: List[Episode]

@dataclass
class ContentItem:
    title: str
    url: str
    poster_url: str
    year: str
    type: str # 'movie' or 'series'
    rating: Optional[str] = None
    description: Optional[str] = None
    seasons: Optional[List[Season]] = None

    def __post_init__(self):
        if self.seasons is None:
            self.seasons = []
