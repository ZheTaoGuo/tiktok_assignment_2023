# Tiktok Immersion Assignment 2023

![Tests](https://github.com/TikTokTechImmersion/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

## Instant Messaging System

### Technologies 
- Architecture: HTTP and RPC Server communicating using RPC IDL
- Data Storage: Redis 

### Prerequisites
- Docker 
- JMeter [Stress Testing Tool]

### Setup
1. git clone https://github.com/ZheTaoGuo/tiktok_assignment_2023.git
2. Ensure you are at the root directory and start the servers 
```bash
docker compose up
```
3. Send a message with the following payload at the endpoint `localhost:8080/api/send` with a `POST` request
```bash
{
  "chat": "Tommy:Jenny",
  "text": "Hey Jenny!"
  "sender": "Tommy"
}
```
4. Receive the message with the following payload at the endpoint `localhost:8080/api/pull` with a `GET` request
```bash
{
  "chat": "Tommy:Jenny"
}
```
