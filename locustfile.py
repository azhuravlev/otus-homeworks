from locust import HttpUser, task, between
import faker

fake = faker.Faker()

class TestUser(HttpUser):
  wait_time = between(5, 15)

  @task(1)
  def get_message(self):
    self.client.get('/messages/' + str(fake.random_int(1, 50)), name='message_id')

  @task(5)
  def search(self):
    self.client.get('/messages?search=' + fake.word(), name='message_search')
