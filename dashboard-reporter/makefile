SCRIPTFILE := report.rb

$(SCRIPTFILE): footer.rb ../lib/dashboard_reporter.rb
	echo "#!/usr/bin/env ruby\n\n" > $(SCRIPTFILE)
	cat ../lib/dashboard_reporter.rb >> $(SCRIPTFILE)
	cat footer.rb >> $(SCRIPTFILE)
	chmod 755 $(SCRIPTFILE)
