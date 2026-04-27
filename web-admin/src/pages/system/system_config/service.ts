import { request } from 'umi';
import type {
  QuerySystemConfigResponse,
  SaveSystemConfigResponse,
  SystemConfigData,
} from './data.d';

export async function querySystemConfig() {
  return request<QuerySystemConfigResponse>('/api/sys/systemConfig/querySystemConfig', {
    method: 'GET',
  });
}

export async function saveSystemConfig(data: SystemConfigData) {
  return request<SaveSystemConfigResponse>('/api/sys/systemConfig/saveSystemConfig', {
    method: 'POST',
    data,
  });
}
