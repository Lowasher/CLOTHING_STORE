-- Создание таблицы ролей
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Добавление базовых ролей из ТЗ
INSERT INTO roles (id, name) VALUES 
('11111111-1111-1111-1111-111111111111', 'Гость'),
('22222222-2222-2222-2222-222222222222', 'Клиент'),
('33333333-3333-3333-3333-333333333333', 'Менеджер'),
('44444444-4444-4444-4444-444444444444', 'Администратор');

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    role_id TEXT,
    registration_date TEXT NOT NULL,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Создание таблицы продуктов
CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT,
    gender TEXT,
    size TEXT,
    color TEXT,
    description TEXT,
    brand TEXT,
    season TEXT,
    price REAL NOT NULL
);

-- Создание таблицы фотографий продуктов
CREATE TABLE IF NOT EXISTS photos (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    product_id TEXT,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Создание таблицы корзины
CREATE TABLE IF NOT EXISTS carts (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    product_id TEXT,
    count INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Создание таблицы заказов
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    datetime TEXT NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Создание таблицы элементов заказа
CREATE TABLE IF NOT EXISTS order_items (
    id TEXT PRIMARY KEY,
    order_id TEXT,
    product_id TEXT,
    price REAL NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
