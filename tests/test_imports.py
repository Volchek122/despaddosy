import sys
import unittest

class TestImports(unittest.TestCase):
    def test_imports(self):
        try:
            import tv_hub.main
            import tv_hub.gui.main_window
            import tv_hub.gui.dashboard
            import tv_hub.gui.flixy_browser
            import tv_hub.gui.flixy_details
            import tv_hub.gui.player
            import tv_hub.gui.flixy_app
            import tv_hub.network.server
            import tv_hub.content.provider
            import tv_hub.content.rezka_provider
            import client.remote_control
        except ImportError as e:
            self.fail(f"Import failed: {e}")

if __name__ == '__main__':
    unittest.main()
