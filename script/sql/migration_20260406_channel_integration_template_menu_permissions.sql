UPDATE sys_menu
SET background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList,/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail,/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate,/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate,/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus',
    update_by = 'codex',
    update_time = NOW()
WHERE id = 314
   OR background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList'
   OR menu_path = '/system/channelIntegrationTemplate/list';
