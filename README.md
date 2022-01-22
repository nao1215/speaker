# speaker - Read the text aloud
speaker command reads aloud the text message. It supports multilingual voice reading. If you want the time signal, the speaker can also start reading aloud at a specified time.  

**The time signal function is being tested.**

# Platform
speaker command depends the cron command to implement the time signal function. So, Speaker command only work UNIX(include Mac) / Linux platform.

# Prerequisite
## macOS
speaker command requies AudioToolbox.framework, but this is automatically linked.

## Linux
ALSA is required. On Ubuntu or Debian, run this command:
```
$ sudo apt install libasound2-dev
```

# How to Install
## Step.1 Install golang
If you don't install golang in your system, please install Golang first. Check [the Go official website](https://go.dev/doc/install) for how to install golang.

## Step2. Install speaker
```
$ go install github.com/nao1215/speaker/cmd/speaker@latest
```

# How to use
## Read the text
Speaker command reads the text in Japanese by default (Because the author is Japanese). 
```
$ speaker "こんにちは"
```

For example, if you want to use russian, please execute the speaker command as follows. Supported languages is listed [here](./doc/SupportedLanguage.md).

```
$ speaker --lang="ru" "я хочу есть мороженое"
```
The speaker command also supports pipes.

```
$ echo "pipe is supported" | speaker --lang="en"
```

## Create mp3 file
If you want to create the mp3 file instead of reading aloud, execute the following command.
```
$ speaker --mp3="./output.mp3" "Create mp3 file"
```

## Register time signal
The time signal is registered by writing the information to the cron configuration file.
```
$ sudo -E speaker -r 17:43 "Register time signal"
```

## Delete time signal
Time signal information is deleted interactively. speaker command display all registered time signals. So, you select one from them, and then speaker command delete it.
```
$ sudo -E ./speaker -d
[1] 43 17 * * * /home/nao/.go/bin/speaker "テストだよ"
[2] 43 17 * * * /home/nao/.go/bin/speaker "Register time signal"
[3] 00 00 * * * /home/nao/.go/bin/speaker "Sweet Dreams"
[4] 00 06 * * * /home/nao/.go/bin/speaker "Good morning"

Which time signal do you delete [1-4]: 2
```

# Contact
If you would like to send comments such as "find a bug" or "request for additional features" to the developer, please use one of the following contacts.

- [GitHub Issue](https://github.com/nao1215/speaker/issues)

# LICENSE
The speaker project is licensed under the terms of the Apache License 2.0.