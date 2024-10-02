У вас должно быть установлено следующее:
-goose
-go v1.22
-psql
-task
1) Установите следующие значения в командной строке:
      - export GOOSE_DRIVER=postgres
      - export GOOSE_DBSTRING=postgres://postgres:postgres@localhost/effective_mobile?sslmode=disable
2) Запустите task файл (https://taskfile.dev/installation/)
      - task init
      - task launch
3) Ознакомиться с документацией можно будет по этой ссылке после запуска приложения: http://localhost:8080/swagger/index.html
