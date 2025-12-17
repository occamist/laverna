# when run inside Anki, __name__ variable would be numeric value (addon ID)
if __name__ != "addon":
    import atexit
    import logging
    import threading
    from pathlib import Path

    from aqt import mw
    from flask import Flask, jsonify, request, Response
    from waitress.server import create_server

    # Configure logging to file only, Anki shows stderr as error popups!
    addon_dir = Path(__file__).parent
    log_file = addon_dir / "addon.log"
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s.%(msecs)03d | %(levelname)s | %(name)s | %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
        handlers=[logging.FileHandler(log_file, mode="a")],
        force=True,
    )
    logger = logging.getLogger(__name__)
    logger.info("addon initializing...")

    app = Flask(__name__)

    @app.before_request
    def log_request_info() -> None:
        logger.info(f"{request.method} {request.path}")

    @app.route("/")
    def index() -> Response:
        return jsonify({"status": "healthy"})

    @app.route("/cards/count")
    def card_count() -> Response:
        count = mw.col.card_count()
        return jsonify({"card_count": count})

    srv = create_server(app, host="127.0.0.1", port=5000)
    def run_srv() -> None:
        logger.info("HTTP server running on http://127.0.0.1:5000")
        srv.run()

    def shutdown_srv() -> None:
        srv.close()
        logger.info("HTTP server shutdown successfully")
    atexit.register(shutdown_srv)

    th = threading.Thread(target=run_srv, daemon=True)
    th.start()
    logger.info("addon initialized successfully")
