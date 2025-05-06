-- Migration: 002_seed_data
-- Description: Insert initial data for testing

-- Insert admin user (password: admin123)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES ('admin', 'admin@example.com', '$2a$10$XEh5REMQQjGLWXgGvB4xU.xvg0lqfQ7Ll0mUYN0oIiYX4zlrEQA0y', 'Admin User', 'admin');

-- Insert regular user (password: user123)
INSERT INTO users (username, email, password_hash, full_name, role)
VALUES ('user', 'user@example.com', '$2a$10$JDDRQEcZ.L8F1Y/x4Juy1uhJA13TxjWWhiG3F2tMTZO44aYhGrSZK', 'Regular User', 'user');

-- Insert categories
INSERT INTO categories (name, description)
VALUES 
    ('Electronics', 'Electronic devices and accessories'),
    ('Clothing', 'Apparel and fashion items'),
    ('Books', 'Books and publications'),
    ('Home & Kitchen', 'Home and kitchen products'),
    ('Sports & Outdoors', 'Sports equipment and outdoor gear');

-- Insert products
INSERT INTO products (name, description, price, stock_quantity, status)
VALUES
    ('Smartphone X', 'Latest smartphone with advanced features', 899.99, 50, 'active'),
    ('Laptop Pro', 'Professional laptop for developers', 1299.99, 25, 'active'),
    ('Wireless Headphones', 'Noise-cancelling wireless headphones', 199.99, 100, 'active'),
    ('T-shirt Basic', 'Cotton basic t-shirt', 19.99, 200, 'active'),
    ('Jeans Classic', 'Classic blue jeans', 49.99, 150, 'active'),
    ('Programming 101', 'Introduction to programming', 29.99, 75, 'active'),
    ('Design Patterns', 'Book about software design patterns', 39.99, 30, 'active'),
    ('Coffee Maker', 'Automatic coffee maker', 89.99, 40, 'active'),
    ('Kitchen Knife Set', 'Professional kitchen knife set', 129.99, 35, 'active'),
    ('Running Shoes', 'Professional running shoes', 79.99, 80, 'active');

-- Associate products with categories
INSERT INTO product_categories (product_id, category_id)
VALUES
    (1, 1), -- Smartphone -> Electronics
    (2, 1), -- Laptop -> Electronics
    (3, 1), -- Headphones -> Electronics
    (4, 2), -- T-shirt -> Clothing
    (5, 2), -- Jeans -> Clothing
    (6, 3), -- Programming book -> Books
    (7, 3), -- Design Patterns book -> Books
    (8, 4), -- Coffee Maker -> Home & Kitchen
    (9, 4), -- Knife Set -> Home & Kitchen
    (10, 5); -- Running Shoes -> Sports & Outdoors

-- Insert some reviews
INSERT INTO reviews (product_id, user_id, rating, comment)
VALUES
    (1, 2, 5, 'Great smartphone, very fast and reliable.'),
    (1, 1, 4, 'Good phone but battery could be better.'),
    (2, 2, 5, 'Perfect laptop for development work.'),
    (3, 1, 4, 'Good sound quality but a bit expensive.'),
    (6, 2, 5, 'Excellent book for beginners.'),
    (8, 1, 3, 'Works well but sometimes leaks.');

-- Add some products to wishlists
INSERT INTO wishlist (user_id, product_id)
VALUES
    (1, 3),
    (1, 5),
    (1, 8),
    (2, 2),
    (2, 7),
    (2, 9); 