# Laverna Addon for Anki Application

This Addon is a bridge server between Laverna CLI and Anki. It's purpose is to act as a middleman to receive the data from Laverna CLI.

Why do we need this Addon? Because Anki doesn't have a SDK so we use this Addon to make new decks for your profile.

Laverna Anki Addon works with Laverna CLI to create custom decks for your needs with Google Translate audio.

- Download Laverna CLI
- Activate Laverna Anki Addon (this addon runs HTTP server on localhost with port 5555)
- Now you can use Laverna CLI to create your decks, for more info, check via `laverna anki --help`

Click here to download it, [Laverna Anki Addon](https://ankiweb.net/shared/info/1504755463?cb=1766308844581) 

## Examples

```sh
laverna anki --profile Talha --deck my-thai-deck --voice th --file ./testdata/anki-th-example.csv
```
