import { request } from 'umi';
import type {
  ChannelIntegrationTemplateCreateResponse,
  ChannelIntegrationTemplateDetailResponse,
  ChannelIntegrationTemplateListParams,
  ChannelIntegrationTemplateListResponse,
  ChannelIntegrationTemplateStatusPayload,
  ChannelIntegrationTemplateSubmitPayload,
} from './data.d';

export async function queryChannelIntegrationTemplateList(
  params: ChannelIntegrationTemplateListParams,
) {
  return request<ChannelIntegrationTemplateListResponse>(
    '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList',
    {
      method: 'GET',
      params: {
        ...params,
      },
    },
  );
}

export async function queryChannelIntegrationTemplateDetail(id: number) {
  return request<ChannelIntegrationTemplateDetailResponse>(
    '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail',
    {
      method: 'GET',
      params: {
        id,
      },
    },
  );
}

export async function createChannelIntegrationTemplate(
  params: ChannelIntegrationTemplateSubmitPayload,
) {
  return request<ChannelIntegrationTemplateCreateResponse>(
    '/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate',
    {
      method: 'POST',
      data: {
        ...params,
      },
    },
  );
}

export async function updateChannelIntegrationTemplate(
  params: ChannelIntegrationTemplateSubmitPayload,
) {
  return request('/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateChannelIntegrationTemplateStatus(
  params: ChannelIntegrationTemplateStatusPayload,
) {
  return request('/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}
