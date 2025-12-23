from concurrent.futures import Future
from contextlib import contextmanager
from http import HTTPStatus
import logging
import os
from pathlib import Path
import tempfile
from threading import Thread
from typing import Iterator

from anki.collection import (
    Collection,
    NotetypeDict,
    ImportCsvRequest,
    Delimiter,
    ImportLogWithChanges,
    CsvMetadata,
)
from anki.consts import MODEL_CLOZE
from anki.stdmodels import StockNotetypeKind
from aqt import mw, gui_hooks, appVersion
from flask import Flask, jsonify, request, Response, Blueprint
from waitress.server import create_server


LEGACY_MODEL_NAME = "Cloze Multi Choice Audio"
MODEL_NAME = "Laverna Cloze"

DEFAULT_ADDRESS = "127.0.0.1"
DEFAULT_PORT = 5555


@contextmanager
def temp_csv_file(csv_data: str) -> Iterator[str]:
    tmp = tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".csv", newline="")
    try:
        tmp.write(csv_data)
        tmp.close()
        yield tmp.name
    finally:
        os.remove(tmp.name)


def save_model(col: Collection) -> NotetypeDict:
    model = col.models.new(MODEL_NAME)
    model["type"] = MODEL_CLOZE
    model["originalStockKind"] = StockNotetypeKind.KIND_CLOZE

    col.models.add_field(model, col.models.new_field("Text"))
    col.models.add_field(model, col.models.new_field("HelperText"))
    col.models.add_field(model, col.models.new_field("TextA"))
    col.models.add_field(model, col.models.new_field("TextB"))
    col.models.add_field(model, col.models.new_field("TextC"))
    col.models.add_field(model, col.models.new_field("TextD"))
    col.models.add_field(model, col.models.new_field("AudioA"))
    col.models.add_field(model, col.models.new_field("AudioB"))
    col.models.add_field(model, col.models.new_field("AudioC"))
    col.models.add_field(model, col.models.new_field("AudioD"))
    col.models.add_field(model, col.models.new_field("AudioAnswer"))

    templ = col.models.new_template("Cloze")
    templ["qfmt"] = """{{cloze:Text}}<br>
{{HelperText}} <br><br>

{{AudioA}} {{TextA}}
<br>
{{AudioB}} {{TextB}}
<br>
{{AudioC}} {{TextC}}
<br>
{{AudioD}} {{TextD}}
"""
    templ["afmt"] = """{{cloze:Text}}<br>
{{HelperText}}

<hr id=answer>

{{AudioAnswer}}"""
    templ["bqfmt"] = "{{cloze:Text}}"
    templ["bafmt"] = "{{cloze:Text}}"
    col.models.add_template(model, templ)

    model["css"] = """.card {
    font-family: 'Sarabun', Arial, sans-serif;
    font-size: 32px;
    text-align: center;
    min-height: 400px;
    padding: 16px;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0,0,0,0.2);
}

.cloze {
    font-weight: bold;
    color: #ffd700;
    font-size: 32px;
}

hr#answer {
    height: 8px;
    background: #00add8;
}"""

    col.models.add(model)
    return col.models.by_name(MODEL_NAME)


# when run inside Anki, __name__ variable would be numeric value (addon ID)
if __name__ != "addon":
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

    cfg = mw.addonManager.getConfig(__name__) or {}
    if not cfg:
        logger.fatal("config is empty")

    app = Flask(__name__)

    @app.before_request
    def log_request_info() -> None:
        logger.info(f"{request.method} {request.path}")

    @app.route("/")
    def index() -> tuple[Response, HTTPStatus]:
        res: dict = {
            "status": "healthy",
            "minimum_required_anki_app_version": appVersion,
        }
        return jsonify(res), HTTPStatus.OK

    # API v1 Blueprint with version prefix
    api_v1 = Blueprint("api_v1", __name__, url_prefix="/v1")

    @api_v1.route("/import-csv", methods=["POST"])
    def import_csv() -> tuple[Response, HTTPStatus]:
        if request.content_type != "text/csv":
            return jsonify(
                {"message": f"Content-type '{request.content_type}' must be 'text/csv'"}
            ), HTTPStatus.BAD_REQUEST

        profile: str | None = request.args.get("profile", default=None)
        if profile is None or profile.strip() == "":
            return jsonify(
                {"message": "Missing 'profile' query parameter"}
            ), HTTPStatus.BAD_REQUEST

        deck_name: str | None = request.args.get("deck", default=None)
        if deck_name is None or deck_name.strip() == "":
            return jsonify(
                {"message": "Missing 'deck' query parameter"}
            ), HTTPStatus.BAD_REQUEST

        raw: str = request.data.decode()
        future: Future = Future()

        def execute() -> None:
            if profile not in mw.pm.profiles():
                return future.set_result((None, f"Profile '{profile}' does not exist"))

            current_profile = mw.pm.name
            if current_profile != profile:
                if mw.col:
                    mw.col.close()
                    mw.col = None
                mw.pm.load(profile)
                mw.loadProfile()
                mw.reset()
                mw.deckBrowser.show()

            col = mw.col
            if col is None:
                return future.set_result((None, "Failed to load collection"))

            model: NotetypeDict | None = col.models.by_name(LEGACY_MODEL_NAME)
            if model is None:
                if MODEL_NAME in col.models.all_names():
                    model = col.models.by_name(MODEL_NAME)
                else:
                    model = save_model(col)

            deck_id = col.decks.id(deck_name, create=True)

            with temp_csv_file(raw) as path:
                # CsvMetadata PB defined here https://github.com/ankitects/anki/blob/main/proto/anki/import_export.proto#L148
                md: CsvMetadata = col.get_csv_metadata(
                    path=path, delimiter=Delimiter.COMMA
                )
                md.deck_id = deck_id
                md.global_notetype.id = model["id"]
                md.global_notetype.field_columns[:] = list(
                    range(1, len(model["flds"]) + 1)
                )
                md.tags_column = 0  # no tags column
                md.dupe_resolution = CsvMetadata.DupeResolution.UPDATE
                md.match_scope = CsvMetadata.MatchScope.NOTETYPE_AND_DECK
                req: ImportCsvRequest = ImportCsvRequest(path=path, metadata=md)
                resp: ImportLogWithChanges = col.import_csv(req)
                mw.reset()
                mw.deckBrowser.show()
                res = {
                    "found_notes": resp.log.found_notes,
                    "updated_notes": len(list(resp.log.updated)),
                    "new_notes": len(list(resp.log.new)),
                }
            future.set_result((res, None))

        # run on main thread of Anki to avoid SQLite threading issues
        mw.taskman.run_on_main(execute)

        # block & wait for the result
        (res, err) = future.result()
        if err is not None:
            return jsonify({"message": err}), HTTPStatus.INTERNAL_SERVER_ERROR
        return jsonify(res), HTTPStatus.OK

    address: str = cfg.get("address", DEFAULT_ADDRESS)
    port: int = cfg.get("port", DEFAULT_PORT)

    app.register_blueprint(api_v1)
    srv = create_server(app, host=address, port=port)

    def run_srv() -> None:
        logger.info(f"HTTP server running on http://{address}:{port}")
        srv.run()

    def shutdown_srv() -> None:
        srv.close()
        logger.info("HTTP server shutdown successfully")

    gui_hooks.profile_will_close.append(shutdown_srv)

    th = Thread(target=run_srv, daemon=True)
    th.start()
    logger.info("addon initialized successfully")
