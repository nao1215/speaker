#!/bin/bash -eu

echo "仕事用の時報を追加します"
sudo -E speaker -r 09:30 "今日も順調  明日も順調"
sudo -E speaker -r 10:00 "コアタイムが始まったよー仕事をしようねー"
sudo -E speaker -r 12:00 "おっひる おっひる おひるのじかんだー"
sudo -E speaker -r 13:00 "おやすみしゅうりょー おやすみしゅうりょー"
sudo -E speaker -r 15:00 "コアタイムが終わったよ これからが本番"
sudo -E speaker -r 16:30 "そろそろ定時  そろそろ定時"
sudo -E speaker -r 21:00 "オネムの時間だー 帰れ帰れ"
echo "追加完了"
