- Вынесен слой диалогов в отдельный микросервис dialog
- В основном сервисе добавлены две новые ручки под /v2
```shell

curl --location 'http://localhost:8080/api/v2/dialog/3/send' \
--header 'Authorization: ...' \
--header 'Content-Type: application/json' \
--data '{
    "text": "Давай дружить"
}'



curl --location 'http://localhost:8080/api/v2/dialog/3/list' \
--header 'Authorization: ...'
```
- межсервисное взаимодейстие по gRPC
- старые ручки под /v1 остались не тронутыми, работают напрямую со UseCase диалогов
- так как микросервис живет в том же домене, что и основной сервис - база осталась общая, но после отключения основных /v1 
- стоит рассмотреть выделение отдельного инстанса база для микросервиса 

Запуск 
```shell
make run-docker
```