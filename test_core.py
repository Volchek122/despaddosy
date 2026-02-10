import unittest
from tv_hub.core.mock import MockProvider
from tv_hub.core.rezka import RezkaProvider
from tv_hub.core.models import ContentItem

class TestCoreLogic(unittest.TestCase):
    def test_mock_search(self):
        provider = MockProvider()
        results = provider.search("Stranger Things")
        self.assertTrue(len(results) > 0)
        self.assertEqual(results[0].title, "Stranger Things")
        print("Mock Search Result:", results[0].title)

    def test_mock_details(self):
        provider = MockProvider()
        item = provider.get_details("mock_url_1")
        self.assertIsNotNone(item.seasons)
        self.assertEqual(len(item.seasons), 2)
        print("Mock Details Seasons:", len(item.seasons))

    def test_rezka_provider_init(self):
        # Just check if it initializes without error (even if API is missing)
        try:
            provider = RezkaProvider()
            self.assertIsNotNone(provider)
            print("RezkaProvider initialized successfully")
        except Exception as e:
            self.fail(f"RezkaProvider failed to initialize: {e}")

if __name__ == '__main__':
    unittest.main()
