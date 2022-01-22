//
// speaker - Read the text aloud
//
// speaker command reads aloud the text message. It supports
// multilingual voice reading. If you want the time signal,
// the speaker can also start reading aloud at a specified time.
//
// speaker command reads the text in Japanese by default (Because
// the author is Japanese).
// $ speaker "こんにちは"
//
// For example, if you want to use russian, please execute the
// speaker command as follows.
// $ speaker --lang="ru" "я хочу есть мороженое"
//
// The speaker command also supports pipes.
// $ echo "pipe is supported" | speaker --lang="en"
//
//
// If you want to create the mp3 file instead of reading aloud,
// execute the following command.
// $ speaker --mp3="./output.mp3" "Create mp3 file"
//
// If you want to register the time signal, the time signal is
// registered by writing the information to the cron configuration file.
// $ sudo -E speaker -r 17:43 "Register time signal"
//
// Time signal information is deleted interactively. speaker command display
// all registered time signals. So, you select one from them, and then speaker
// command delete it.
// $ sudo -E ./speaker -d
// [1] 43 17 * * * /home/nao/.go/bin/speaker "テストだよ"
// [2] 43 17 * * * /home/nao/.go/bin/speaker "Register time signal"
// [3] 00 00 * * * /home/nao/.go/bin/speaker "Sweet Dreams"
// [4] 00 06 * * * /home/nao/.go/bin/speaker "Good morning"
//
// Which time signal do you delete [1-4]: 2
package main
