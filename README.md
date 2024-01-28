# 老黄历精简版Go

[![Chat on Telegram](https://img.shields.io/badge/@LuckyYUI_bot-2CA5E0.svg?logo=telegram&label=Telegram)](https://t.me/LuckyYUI_bot)
![GitHub Repo stars](https://img.shields.io/github/stars/Nigh/laohuangli-lite-go?style=flat&color=ffaaaa)
[![Software License](https://img.shields.io/github/license/Nigh/laohuangli-lite-go)](LICENSE)
![Docker](https://img.shields.io/badge/Build_with-Docker-ffaaaa)

这个项目是对 [青年老黄历Bot](https://github.com/HerbertGao/laohuangli_bot) 项目的精简复刻。仅保留了每日老黄历功能。并增加了提名新词条功能，可以通过投票添加新的词条。

## 部署

1. 在 `.env` 中设置你的bot token
2. 根据需要运行下面的命令

```shell
# 初次运行
make
# 升级容器
make upgrade
# 移除容器
make clean
```
