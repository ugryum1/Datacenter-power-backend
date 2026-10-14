# Проектирование систем и продуктовая веб-разработка 2026 (Разработка Интернет Приложений)

### Тема: Расчитать объем потребляемой датацентром электроэнергии в зависимости от оборудования в нем. Услуги - виды устройств (серверы, коммутаторы, СХД и тд), заявка - расчет потреблямой мощности с учетом охлаждения для датацентра со списком оборудования

## API

Запуск: `docker compose up -d`, `go run ./cmd/migrate`, `go run ./cmd/datacenter-power`. Все методы под `http://localhost:8080/api`. Swagger: `http://localhost:8080/swagger/index.html`.

Тело ответа «устройство»:

```json
{
  "id": 16,
  "name": "Ленточная библиотека IBM",
  "short_description": "Роботизированная ленточная библиотека для резервного копирования",
  "image_url": "http://localhost:9000/datacenter-media/datacenter-device-16-image.jpg",
  "video_url": "http://localhost:9000/datacenter-media/datacenter-device-16-video.webm",
  "power_consumption_w": 1500,
  "heat_dissipation_btu": 5118,
  "likes_count": 1
}
```

### Устройства

| Метод | URL | Тело/параметры | Ответ |
|---|---|---|---|
| GET | `/datacenter-equipment` | `?power=N`: если > 0, мощность от N Вт | 200, массив опубликованных устройств; у каждого дополнительно `is_creator`: 1, если его создал текущий инженер, иначе 0 |
| GET | `/datacenter-equipment/reel` | — | 200, первое опубликованное устройство с `is_creator` |
| GET | `/datacenter-equipment/reel/:id` | `?next=true`: следующее после id, после последнего снова первое | 200, устройство с `is_creator`; 404 |
| GET | `/datacenter-equipment/draft` | — | 200, черновик текущего инженера; 404, если нет |
| POST | `/datacenter-equipment` | form-data: `name`, `image` (jpeg/png/webp), `video` (webm/mp4) | 201, устройство; 400, если черновик уже есть или данные некорректны |
| PUT | `/datacenter-equipment/:id/publish` | JSON, все обязательны: `name`, `short_description`, `power_consumption_w`, `heat_dissipation_btu` | 200, устройство; 404, если не черновик текущего инженера |
| DELETE | `/datacenter-equipment/:id` | — (логическое удаление опубликованного устройства текущего инженера) | 200 без тела; 404, если устройство чужое или не опубликовано |
| POST | `/datacenter-equipment/:id/like` | JSON `{"like": 1}` ставит, `{"like": 0}` снимает | 200, устройство; 404, если не опубликовано |

### Инженеры

| Метод | URL | Тело/параметры | Ответ |
|---|---|---|---|
| POST | `/datacenter-engineers/sign-up` | JSON `{login, password}` | 201, `{id, login, is_moderator}`; 400, если логин занят |
| POST | `/datacenter-engineers/sign-in` | JSON `{login, password}` | 200, `{expires_in, access_token, token_type}`; 403 при неверном логине или пароле |
| POST | `/datacenter-engineers/sign-out` | — | 200 без тела, токен попадает в блеклист Redis |

## Таблицы БД

### datacenter_devices — устройства

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint | первичный ключ |
| `name` | varchar(100) | название |
| `short_description` | varchar(500) | краткое описание, заполняется при публикации |
| `status` | varchar(15) | черновик, опубликован или удалён |
| `image_url` | varchar(255) | ссылка на фото в MinIO |
| `video_url` | varchar(255) | ссылка на видео в MinIO |
| `power_consumption_w` | integer | потребляемая мощность, Вт |
| `heat_dissipation_btu` | integer | тепловыделение, БТЕ/ч |
| `date_create` | timestamptz | дата создания |
| `creator_id` | bigint | создатель, внешний ключ на `datacenter_engineers` |
| `date_form` | timestamptz | дата публикации |

### datacenter_engineers — инженеры

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint | первичный ключ |
| `login` | varchar(25) | логин, уникальный |
| `password` | varchar(100) | SHA-1 хэш пароля |
| `is_moderator` | boolean | признак модератора |

### datacenter_device_likes — лайки

| Поле | Тип | Описание |
|---|---|---|
| `id` | bigint | первичный ключ |
| `engineer_id` | bigint | кто поставил лайк, внешний ключ на `datacenter_engineers` |
| `device_id` | bigint | какому устройству, внешний ключ на `datacenter_devices` |
