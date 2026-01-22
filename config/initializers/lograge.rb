class CustomLogFormatter < Lograge::Formatters::KeyValue
  def call(data)
    level = log_level(data[:status])
    "#{Time.current.strftime('%Y-%m-%d %H:%M:%S.%L')} [#{level}] #{super(data)}"
  end

  private

  def log_level(status)
    case status
    when 200..399 then "INFO"
    when 400..499 then "WARN"
    when 500..599 then "ERROR"
    else "INFO"
    end
  end
end

Rails.application.configure do
  config.lograge.enabled = true
  config.lograge.formatter = CustomLogFormatter.new
end
