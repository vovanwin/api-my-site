Api сервис сделаный на основе echo pagoda, но переделаный под использование api.

Бэкенд
- Echo : высокопроизводительный, расширяемый, минималистичный веб-фреймворк Go.
- Ent : простая, но мощная ORM для моделирования и запроса данных.

Хранилище
- PostgreSQL : самая передовая в мире реляционная база данных с открытым исходным кодом.
- Redis : хранилище структур данных в памяти, используемое в качестве базы данных, кеша и брокера сообщений.

## Запустить приложение
После проверки репозитория из корня запустите контейнеры Docker для базы данных и кеша, выполнив make up:

```
git clone git@github.com:vovanwin/api-my-site.git
cd pagoda
make up
```

Поскольку этот репозиторий является шаблоном , а не библиотекой Go , вы не используете файлы go get.
Как только это будет завершено, вы можете запустить приложение, выполнив make run. 
По умолчанию вы должны иметь доступ к приложению в своем браузере по адресу `localhost:8000`.
Если вы когда-нибудь захотите быстро удалить контейнеры Docker и 
перезапустить их, чтобы стереть все данные, выполните команду `make reset`