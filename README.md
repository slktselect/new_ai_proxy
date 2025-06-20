# AI API Proxy

vscode 即使配置了代理 当使用roo code插件来写代码的时候还是走的未配置代理前的网络
用这个项目做roo code的代理
在roo code 配置里面改成本地地址和你配置的端口号
在docker-compose.yaml 中配置代理和启动的端口号

支持gemini

## 使用 Docker Compose

```bash
docker-compose up --build
```

## 贡献

欢迎贡献！请提交 Pull Request 或报告 Issue。

## 许可证

本项目采用 MIT 许可证。