##  tcp-graceful

一个基于golang的优雅热重启库实现，支持golang-TCP服务

关联文章：[Golang 系统编程：如何实现对后台服务优雅的热重启？](https://pandaychen.github.io/2021/11/20/A-GRACEFUL-RESTART-IN-SSHD/)


测试方法：
1.  启动主程序，通过`cat /var/run/tmp.pid`获取当前主进程的`$pid`
2.  通过`kill -SIGHUP $pid`执行主程序的热重启机制
3.  重启后的生成新的主程序不改变当前运行的相对路径