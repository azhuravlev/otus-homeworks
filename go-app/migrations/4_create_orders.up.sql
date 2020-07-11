CREATE TABLE `orders` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `product_id` int(11) NOT NULL,
  `count` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `created_at` datetime DEFAULT NOW(),
  `updated_at` datetime DEFAULT NOW() ON UPDATE NOW(),
  PRIMARY KEY (`id`),
  INDEX (user_id)
) DEFAULT CHARSET=utf8;