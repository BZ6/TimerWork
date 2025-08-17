# Timer Work - Система учета рабочего времени

Полнофункциональная система для отслеживания рабочего времени с установкой целей, визуализацией прогресса и подробной историей работы.

## 🚀 Основные возможности

### 🔐 Управление пользователями
- **Регистрация и авторизация** с JWT токенами
- **Безопасность** - bcrypt хеширование паролей
- **Индивидуальный учет** для каждого пользователя

### ⏱️ Управление рабочим временем
- **Начало недели** - запуск нового таймера с установкой цели
- **Конец недели** - остановка и сохранение времени
- **Пауза/Продолжение** - управление таймером без потери данных
- **Ручное обновление** времени по требованию
- **Точный учет** - время пауз не включается в рабочее время

### 🎯 Система целей
- **Установка цели** при старте недели (от 1 минуты до любого количества часов)
- **Гибкие единицы** - часы или минуты
- **Визуальный прогресс** - прогресс-бар в реальном времени
- **Цветовая индикация** достижения целей
- **Процент выполнения** цели

### 📊 История и аналитика
- **Навигация** между страницами "Таймер" и "История недель"
- **Детальная таблица** всех рабочих недель с колонками:
  - № Недели
  - Начало работы
  - Конец работы  
  - Статус (Активна, Завершена, Приостановлена, Не начата)
  - Продолжительность
  - Цель / Прогресс с мини-прогресс-барами
- **Сортировка** по дате (новые сверху)
- **Лимит** 50 последних недель

### 🎨 Пользовательский интерфейс
- **Современный дизайн** с адаптивной версткой
- **Интуитивные кнопки** с цветовым кодированием
- **Модальные окна** для важных действий
- **Статусные индикаторы** с разными цветами
- **Hover-эффекты** и плавные анимации

## 🛠 Технологический стек

### Backend
- **Go 1.21** с Gin framework
- **PostgreSQL 15** база данных с retry логикой подключения
- **JWT** авторизация с middleware
- **bcrypt** безопасное хеширование паролей
- **CORS** поддержка для cross-origin запросов

### Frontend
- **React 18** с Create React App
- **Axios** для API запросов с interceptors
- **CSS3** с flexbox и grid layouts
- **ES6+** современный JavaScript

### DevOps
- **Docker** multi-stage builds для оптимизации
- **Docker Compose** для оркестрации сервисов
- **Hot reload** в development режиме
- **Environment variables** для конфигурации

## 🚀 Быстрый старт

### Предварительные требования
- Docker и Docker Compose
- Порты 3000, 8080, 5432 должны быть свободны (development)
- Порты 80, 8080, 5432 должны быть свободны (production)

### Development режим

```bash
# Клонирование репозитория
git clone https://github.com/BZ6/TimerWork.git
cd TimerWork

# Запуск development версии
docker-compose up --build
```

### Production режим

```bash
# Клонирование репозитория
git clone https://github.com/BZ6/TimerWork.git
cd TimerWork

# Запуск production версии
docker-compose -f docker-compose.prod.yml up --build -d
```

### Различия между режимами

| Параметр | Development | Production |
|----------|-------------|------------|
| Frontend | React dev server | Nginx + React build |
| Порт фронтенда | 3000 | 80 |
| Оптимизация | Нет | Да |
| Hot reload | Да | Нет |
| Gzip сжатие | Нет | Да |
| Кэширование статики | Нет | Да |
| Размер образа | Больше | Меньше |
| Время запуска | Быстрее | Медленнее (сборка) |

### Доступ к приложению

**Development:**
- **Frontend**: <http://localhost:3000>
- **Backend API**: <http://localhost:8080/api>
- **PostgreSQL**: localhost:5432

**Production:**
- **Frontend**: <http://YOUR_SERVER_IP>
- **Backend API**: <http://YOUR_SERVER_IP:8080/api>
- **PostgreSQL**: localhost:5432 (только внутри Docker network)

## 📡 API Documentation

### Публичные endpoints
```http
POST /api/register         # Регистрация пользователя
POST /api/login            # Вход в систему
```

### Защищенные endpoints (требуют Authorization: Bearer <token>)
```http
GET  /api/workweek              # Получить текущую рабочую неделю
GET  /api/workweek/history      # Получить историю рабочих недель  
POST /api/workweek/start        # Начать новую рабочую неделю с целью
POST /api/workweek/end          # Завершить рабочую неделю
POST /api/workweek/pause        # Поставить таймер на паузу
POST /api/workweek/resume       # Возобновить таймер
GET  /api/workweek/current-time # Получить текущее рабочее время
```

### Примеры запросов

#### Начало недели с целью
```bash
curl -X POST http://localhost:8080/api/workweek/start \
  -H "Authorization: Bearer <your-jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"goal_minutes": 2400}'  # 40 часов
```

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user123", "password": "securepassword"}'
```

## 🗄️ Схема базы данных

### Таблица `users`
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Таблица `work_weeks`
```sql
CREATE TABLE work_weeks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    week_start TIMESTAMP NOT NULL,
    week_end TIMESTAMP,
    last_update_time TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'stopped',
    pause_start TIMESTAMP,
    total_pause_time INTEGER DEFAULT 0,
    week_goal_minutes INTEGER DEFAULT 2400,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 🎮 Руководство пользователя

### Первое использование
1. **Регистрация** - создайте аккаунт с уникальным именем пользователя
2. **Вход** - войдите в систему с вашими учетными данными
3. **Начало недели** - нажмите "Начало недели" и установите цель
4. **Отслеживание** - используйте кнопки управления таймером
5. **История** - переключитесь на "История недель" для просмотра статистики

### Работа с целями
- **Установка**: При старте недели выберите количество и единицы (часы/минуты)
- **Отслеживание**: Прогресс отображается в реальном времени
- **Достижение**: При достижении 100% прогресс-бар становится зеленым

### Управление временем
- **Пауза**: Временно останавливает таймер (время паузы не засчитывается)
- **Продолжение**: Возобновляет работу таймера
- **Обновление**: Ручное обновление отображаемого времени
- **Конец недели**: Финализирует неделю и сохраняет результат

## 🔧 Разработка

### Локальная разработка
```bash
# Запуск только PostgreSQL
docker-compose up postgres

# Backend разработка
cd backend
go mod download
go run main.go

# Frontend разработка
cd frontend
npm install
npm start
```

### Переменные окружения
```env
# Backend
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=timerwork
JWT_SECRET=your-secret-key

# Frontend
REACT_APP_API_URL=http://localhost:8080/api
```

### Управление контейнерами
```bash
# Остановка всех сервисов
docker-compose down                              # development
docker-compose -f docker-compose.prod.yml down  # production

# Просмотр логов
docker-compose logs -f                           # development
docker-compose -f docker-compose.prod.yml logs -f  # production

# Перезапуск отдельного сервиса
docker-compose restart frontend                 # development
docker-compose -f docker-compose.prod.yml restart frontend  # production
```

## 🔒 Безопасность

- **JWT токены** с expiration time
- **bcrypt** хеширование паролей (cost 10)
- **CORS** настройки для безопасности
- **SQL injection** защита через параметризованные запросы
- **Input validation** на frontend и backend

## 📈 Производительность

- **Connection pooling** для PostgreSQL
- **JWT middleware** с кэшированием
- **Эффективные SQL запросы** с индексами
- **Gzip compression** для статических файлов
- **Multi-stage Docker builds** для оптимизации размера

## 🐛 Отладка

### Логи контейнеров
```bash
# Все логи
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres
```

### Подключение к базе данных
```bash
docker-compose exec postgres psql -U postgres -d timerwork
```

### Проверка API
```bash
# Health check
curl http://localhost:8080/api/workweek

# С авторизацией
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/workweek
```

### Проблемы с запуском

#### Фронтенд недоступен по IP
1. **Проверьте порты**: убедитесь что порт 80 (production) или 3000 (dev) открыт
2. **Firewall**: настройте файрвол для входящих соединений
3. **Bind адрес**: проверьте что контейнеры привязаны к 0.0.0.0
4. **Логи**: `docker-compose logs frontend` или `docker-compose -f docker-compose.prod.yml logs frontend`

#### Backend API недоступен
1. **CORS настройки**: убедитесь что CORS разрешает запросы с вашего домена
2. **Сеть**: проверьте что backend и frontend в одной Docker сети
3. **Environment**: проверьте переменные окружения
4. **Логи**: `docker-compose logs backend` или `docker-compose -f docker-compose.prod.yml logs backend`

#### База данных не подключается
1. **Порты**: проверьте что PostgreSQL контейнер запущен
2. **Переменные**: убедитесь что DB_HOST, DB_USER, DB_PASSWORD правильные
3. **Сеть**: проверьте подключение между контейнерами
4. **Логи**: `docker-compose logs postgres` или `docker-compose -f docker-compose.prod.yml logs postgres`

### Команды для диагностики
```bash
# Статус контейнеров
docker-compose ps                                # development
docker-compose -f docker-compose.prod.yml ps    # production

# Проверка сетей
docker network ls
docker network inspect timerwork_app-network

# Проверка портов (development)
netstat -tulpn | grep :3000
netstat -tulpn | grep :8080

# Проверка портов (production)  
netstat -tulpn | grep :80
netstat -tulpn | grep :8080

# Тест подключения к API
curl -v http://localhost:8080/api/register       # development
curl -v http://YOUR_SERVER_IP:8080/api/register  # production

# Перезапуск конкретного сервиса
docker-compose restart frontend                 # development
docker-compose -f docker-compose.prod.yml restart frontend  # production
```

## 📝 Changelog

### v2.0.0 - Система целей и прогресса
- ✅ Добавлена установка целей при старте недели
- ✅ Визуализация прогресса с прогресс-барами
- ✅ Расширенная история недель с прогрессом
- ✅ Улучшенная навигация между страницами
- ✅ Модальные окна для важных действий

### v1.0.0 - Базовая система учета времени
- ✅ Авторизация пользователей
- ✅ Управление рабочим временем
- ✅ Учет пауз
- ✅ Ручное обновление времени
- ✅ Docker deployment

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature ветку (`git checkout -b feature/AmazingFeature`)
3. Commit изменения (`git commit -m 'Add some AmazingFeature'`)
4. Push в ветку (`git push origin feature/AmazingFeature`)
5. Откройте Pull Request

## 📄 Лицензия

Распространяется под лицензией MIT. См. `LICENSE` для подробностей.

## 📞 Поддержка

Если у вас есть вопросы или проблемы:
- Создайте Issue в репозитории
- Проверьте документацию выше
- Посмотрите логи контейнеров для диагностики

---

**Timer Work** - эффективный инструмент для отслеживания рабочего времени с современным интерфейсом и гибкой системой целей! 🚀
