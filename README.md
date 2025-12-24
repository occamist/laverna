# Laverna

[![Version](https://img.shields.io/github/tag/mrwormhole/laverna.svg)](https://github.com/mrwormhole/laverna/tags)
[![CI](https://github.com/mrwormhole/laverna/actions/workflows/main.yaml/badge.svg)](https://github.com/mrwormhole/laverna/actions/workflows/main.yaml)
[![Godoc](https://godoc.org/github.com/mrwormhole/laverna?status.svg)](https://godoc.org/github.com/mrwormhole/laverna)
[![Report Card](https://goreportcard.com/badge/github.com/mrwormhole/laverna)](https://goreportcard.com/report/github.com/mrwormhole/laverna)
[![License](https://img.shields.io/github/license/mrwormhole/laverna)](https://github.com/mrwormhole/laverna/blob/main/LICENSE)
[![Codecov](https://codecov.io/github/mrwormhole/laverna/graph/badge.svg?token=HUR58C4DFN)](https://codecov.io/github/mrwormhole/laverna)

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
  laverna run --file example.csv
```

or

```shell
  laverna run --file example.yaml 
```

### Anki

Make sure you have [Laverna Anki Addon](https://github.com/mrwormhole/laverna/tree/main/addon) installed first. Then you can proceed with the below CSV format.

```csv
Text,HelperText,TextA,TextB,TextC,TextD
ฉันชอบ{{c1::ฟัง}}เพลง,I like to listen to music,ฟัง,เล่น,ดู,อ่าน
เขา{{c1::เล่น}}ฟุตบอลทุกวัน,He plays football every day,เล่น,ฟัง,ดู,อ่าน
คุณ{{c1::อ่าน}}หนังสือนี้ไหม,Do you read this book,อ่าน,ฟัง,เล่น,ดู
```

```shell
  laverna anki --profile Talha --deck my-thai-deck --voice th --file ./testdata/anki-th-example.csv
```

Now your new deck is created in Anki with the audio.

![anki-front-card](https://github.com/user-attachments/assets/fbc19bd1-0d74-4659-b9a0-2e40e9e8d4cf)

![anki-back-card](https://github.com/user-attachments/assets/6cca4cfc-237a-493f-8306-0972278231fa)

## Shell Completions

Output shell completion script for bash, zsh, fish, or Powershell.
Source the output to enable completion.

### Bash / Zsh

```source <(laverna completion bash)``` or ```source <(laverna completion zsh)```

### Fish

```laverna completion fish > ~/.config/fish/completions/laverna.fish```
