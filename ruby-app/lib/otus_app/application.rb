# frozen_string_literal: true

require File.join(__dir__, '..', 'otus_app.rb')

require 'sinatra'
require 'sinatra/json'

module OtusApp
  class Application < ::Sinatra::Base
    configure do
      set :show_exceptions, false
      enable :logging
      use Rack::CommonLogger, STDOUT
    end

    get '/health/' do
      json status: 'ok'
    end
  end
end
