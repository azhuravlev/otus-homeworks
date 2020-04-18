# frozen_string_literal: true

# Specifies the `port` that Puma will listen on to receive requests, default is 3000.
port        ENV.fetch('APP_PORT') { 3000 }
# Specifies the `environment` that Puma will run in.
environment ENV.fetch('RACK_ENV') { 'production' }

threads_count = ENV.fetch('MAX_THREADS') { 1 }

threads 0, threads_count