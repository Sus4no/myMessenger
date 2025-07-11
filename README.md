## Использование мессенджера (локально)

1. Клонировать репозиторий в любую удобную папку `git clone https://github.com/Sus4no/myMessenger`
2. Замените значения на актуальные в `server.conf`
3. Выполните в папке Server/crypto:
   ```bash
   openssl genpkey -algorithm RSA -out server.key
   openssl req -x509 -new -key server.key -out server.crt -days 365 -config server.conf
4. Скопируйте server.crt в папку Client/crypto 
   ```bash
   mkdir Client/crypto
   cp Server/crypto/server.crt Client/crypto/server.crt
5. Соберите отдельно сервер и клиента
   ```bash
    go build server.go server_networking.go
    go build client.go client_networking.go
5. Запустите сервер
    ```bash
    ./server
6. Запустите клиента(-ов)
    ```bash
    ./cleint
    ```
`help` после запуска клиента покажет список доступных команд
