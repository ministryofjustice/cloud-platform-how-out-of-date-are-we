module OrphanedResources
  class HostedZoneTuple < ResourceTuple
    attr_reader :hosted_zone_id

    HOSTED_ZONE_URL = "https://console.aws.amazon.com/route53/v2/hostedzones#ListRecordSets/"

    def initialize(params)
      super params
      @hosted_zone_id = params.fetch(:hosted_zone_id).sub(/.hostedzone./, "") # /hostedzone/ZDINZWZ0PO44E -> ZDINZWZ0PO44E
    end

    def aws_console_url
      HOSTED_ZONE_URL + hosted_zone_id
    end
  end
end
