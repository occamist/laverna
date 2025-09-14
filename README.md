# Laverna

[![Version](https://img.shields.io/github/tag/mrwormhole/laverna.svg)](https://github.com/mrwormhole/laverna/tags)
[![CI Build](https://github.com/mrwormhole/laverna/actions/workflows/tests.yaml/badge.svg)](https://github.com/mrwormhole/laverna/actions/workflows/tests.yaml)
[![GoDoc](https://godoc.org/github.com/mrwormhole/laverna?status.svg)](https://godoc.org/github.com/mrwormhole/laverna)
[![Report Card](https://goreportcard.com/badge/github.com/mrwormhole/laverna)](https://goreportcard.com/report/github.com/mrwormhole/laverna)
[![License](https://img.shields.io/github/license/mrwormhole/laverna)](https://github.com/mrwormhole/laverna/blob/main/LICENSE)
[![Coverage Status](https://coveralls.io/repos/github/mrwormhole/laverna/badge.svg?branch=main)](https://coveralls.io/github/mrwormhole/laverna?branch=main)

the goddess of the thieves, helps you steal translation speeches from the G-daddy monopoly.

<img src="https://github.com/user-attachments/assets/d1d344c9-f36b-4cf7-af70-f162f93ea9f0" width="400" alt="goddess-of-thieves">

## Installation

### Grab Binaries

You can find the binaries through GitHub [releases](https://github.com/mrwormhole/laverna/releases/).

### Install Via Go

```shell
  go install github.com/mrwormhole/laverna@latest
```

## Sample Usage

### Basic

Let's create example CSV

```csv
speed,voice,text
normal,th,สวัสดีครับ
slower,en,Hello there
slowest,ja,こんにちは~
```

or you could do YAML

```yaml
- speed: normal
  voice: th
  text: "สวัสดีครับ"
- speed: slower
  voice: en
  text: "Hello there"
- speed: slowest
  voice: ja
  text: "こんにちは~"
```

Running below command will generate audios in the same directory.

```shell
  laverna run -file example.csv
```

or

```shell
  laverna run -file example.yaml 
```

### Anki

Make sure you have note type installed for Anki. Here is the [note type](./note-type.apkg), when you imported it, you will see "Cloze Multi Choice Audio" note type in your `Anki > Tools > Manage Note Types`. Then you can proceed with the below Anki CSV format.

```csv
Text,HelperText,TextA,TextB,TextC,TextD
ฉันชอบ{{c1::ฟัง}}เพลง,I like to listen to music,ฟัง,เล่น,ดู,อ่าน
```

```shell
  laverna anki --profile Talha --voice th --file ./testdata/anki-th-example.csv
```

Now you will see a new result as below CSV file, all of your media is generated and inserted into Anki, however in order to create a deck in Anki, you need to import CSV below. And carefully choose comma delimiter with note type "Cloze Multi Choice Audio"

```csv
Text,HelperText,TextA,TextB,TextC,TextD,AudioA,AudioB,AudioC,AudioD,AudioAnswer
ฉันชอบ{{c1::ฟัง}}เพลง,I like to listen to music,ฟัง,เล่น,ดู,อ่าน,[sound:a.mp3],[sound:b.mp3],[sound:c.mp3],[sound:d.mp3],[sound:e.mp3]
```

![anki-front-card](https://github.com/user-attachments/assets/fbc19bd1-0d74-4659-b9a0-2e40e9e8d4cf)

![anki-back-card](https://github.com/user-attachments/assets/6cca4cfc-237a-493f-8306-0972278231fa)

## Shell Completions

Output shell completion script for bash, zsh, fish, or Powershell.
Source the output to enable completion.

### Bash / Zsh

```source <(laverna completion bash)``` or ```source <(laverna completion zsh)```

### Fish

```laverna completion fish > ~/.config/fish/completions/laverna.fish```
