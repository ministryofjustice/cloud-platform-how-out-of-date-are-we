class Filestore
  def list_files
    Dir["data/**/*.json"]
  end

  def store_file(file, content)
    dir = File.dirname(file)
    FileUtils.mkdir_p(dir) unless FileTest.directory?(dir)
    File.write(file, content)
  end

  def retrieve_file(file)
    File.read(file)
  end
end
