# Laverna Addon for Anki

This addon is a bridge server between Laverna CLI and Anki. It's purpose is to act as a middleman to receive the data from Laverna CLI.

Why do we need this addon? Because Anki doesn't have a SDK so we use this addon to make new decks for your profile.

This addon works with Laverna CLI to create custom decks for your needs with Google Translate audio.

- Download Laverna CLI
- Install this addon
- Start Anki (the addon runs HTTP server on localhost with port 5555)
- Now you can use Laverna CLI to create your decks, for more info, see `laverna anki --help` or [the git repository](https://github.com/mrwormhole/laverna)

Click [here](https://ankiweb.net/shared/info/1504755463) to download the addon.

## Check the status of the addon via Curl

```sh
❯ curl http://localhost:5555/
{"minimum_required_anki_app_version":"25.09.2","status":"healthy"}
```
