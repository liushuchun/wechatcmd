# wechatcmd  [![star this repo](http://github-svg-buttons.herokuapp.com/star.svg?user=hawklithm&repo=wechatcmd&style=flat&background=1081C1)](http://github.com/hawklithm/wechatcmd) [![fork this repo](http://github-svg-buttons.herokuapp.com/fork.svg?user=hawklithm&repo=wechatcmd&style=flat&background=1081C1)](http://github.com/hawklithm/wechatcmd/fork) ![Build](https://camo.githubusercontent.com/46cb8b3469febc6cdb6fbaea2ef1517c396004e7/68747470733a2f2f7472617669732d63692e6f72672f736a77686974776f7274682f676f6c6561726e2e706e673f6272616e63683d6d6173746572)

公司出于安全性考虑不允许安装pc版wechat，网页版在使用上并不令人满意，在一番调研之后决定采用 
[liushuchun/wechatcmd](https://github.com/liushuchun/wechatcmd)
，在阅读了源码并了解到原作者已放弃继续开发，遂决定fork一份之后在此基础上继续进行开发，完善其功能

目前已完善点：

- [x] termui版本升级到3.0.0，接口兼容问题修复
- [x] 群聊天中发言人显示
- [x] 用户多端登陆时，通过其他端发出的消息的同步
- [x] 切换当前聊天窗口时，历史聊天记录的恢复
- [x] 干掉了红包提醒(逻辑存在bug，误提醒，让人很烦躁，所以删掉了)

**注：本程序目的为日常使用替代pc端微信，所以不会开发自动回复或者聊天机器人抑或是群发之类的功能**


操作方式：

| 按键 | 说明 |
| --- | --- |
| Ctrl+n | 下一个聊天 |
| Ctrl+p | 上一个聊天 |
| Ctrl+j | 下一条聊天记录 |
| Ctrl+k | 上一条聊天记录 |
| Ctrl+w | 展示选中的聊天记录的图片内容 |
| Ctrl+c | 退出 |

开发计划：

- [x] 实现微信登陆(原版已实现)
- [x] 实现微信认证(原版已实现)
- [x] 实现拉取用户信息(原版已实现)
- [x] 同步消息
- [x] 自动更新消息
- [x] 聊天
- [x] 群聊
- [x] 支持图片显示
- [x] 支持emoji表情
- [ ] 自动保存消息到本地
- [ ] vim式操作
- [ ] 消息提醒
- [ ] 表情包
- [ ] 本地表情包发送(发图片)
- [ ] 解析分享消息
- [ ] 解析公众号消息





以下是原版的README
=================
## 微信命令行版本
开发这个命令行版本，一是为了熟悉微信的接口，二是方便咱们习惯命令行的同学。

现在是中文的支持不是很好，还没有什么特别好的解决方法。

项目还是远未完成，热烈欢迎有兴趣的朋友一起加入开发。

有什么建议可以提issue。谢谢，也欢迎直接提PR。


### 功能特性

1. 用户检索
2. 聊天表情包快捷键
3. 自动聊天
4. Vimer式快捷键让操作丝般顺滑
5. 更加Geek的feel.


### 键盘快捷键


<table>
    <tr><td>Ctrl-n</td><td>下一页</td></tr>
    <tr><td>Ctrl-p</td><td>上一页</td></tr>
    <tr><td>Enter</td><td>输入</td></tr>
    <tr><td>Ctrl-c</td><td>退出</td></tr>
    <tr><td>Ctrl-1</td><td>退出自动回复</td></tr>
    <tr><td>Ctrl-2</td><td>启用自动回复</td></tr>
    <tr><td>Ctrl-3</td><td>机器人自动回复(还没好)</td></tr>
</table>

### 运行bin文件
linux,mac,windows编好的包分别在install 下面的linux/ mac/ win/下。(windows暂时支持的不好，虽然是交叉编译可以运行，但是其UI机制和unix系差的很多，termui支持的并不是很好)

```
git clone git@github.com:liushuchun/wechatcmd.git
cd wechatcmd/install/
进入各自目录
```


### Mac安装

	$ go get -u github.com/hawklithm/wechatcmd


### Linux安装

	$ go get -u github.com/hawklithm/wechatcmd


### 现在实现的界面：

![聊天动态图](https://raw.githubusercontent.com/liushuchun/wechatcmd/master/img/show.gif)
出现二维码之后，使用微信扫描二维码，进行登录。
![登陆后图](https://raw.githubusercontent.com/liushuchun/wechatcmd/master/img/wechatcmd-1.png)
![聊天图片](https://raw.githubusercontent.com/liushuchun/wechatcmd/master/img/wechatcmd-2.png)




### 使用

	$ wechatcmd

### 现在完成的功能
- [x] 实现微信登陆
- [x] 实现微信认证
- [x] 实现拉取用户信息
- [x] 同步消息
- [x] 设置自动回复：正在忙，稍后回来，等等。
- [x] 自动更新消息
- [x] 自动回复消息
- [x] 获取其他消息
- [x] 聊天
- [ ] 群聊
- [ ] 读取图片
- [ ] 自动保存消息到本地
- [ ] 表情包的翻译

### 由于工作太忙，后期已经没有精力继续开发，欢迎有兴趣的同学继续开发
