require "mysql2"
require "faker"

db = Mysql2::Client.new(host: ENV["DB_HOST"], post: ENV["DB_PORT"], username: ENV["DB_USER"], password: ENV["DB_PASS"], database: ENV["DB_NAME"])


1000.times do
  db.query("insert into Messages (subject, body, user_id, created_at, updated_at) values(#{Faker::Book.title}, #{Faker::Lorem.paragraph}, 1, #{Time.utc}, #{Time.utc}")
end
