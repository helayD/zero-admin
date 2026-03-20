import { request } from 'umi';
import type { NoticeListItem, NoticeListParams } from './data.d';

export async function addNotice(params: NoticeListItem) {
  return request('/api/sys/notice/addNotice', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeNotice(ids: number[]) {
  return request(`/api/sys/notice/deleteNotice?ids=${ids.join(',')}`, {
    method: 'GET',
  });
}

export async function updateNotice(params: NoticeListItem) {
  return request('/api/sys/notice/updateNotice', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateNoticeStatus(params: { ids: number[]; status: number }) {
  return request('/api/sys/notice/updateNoticeStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryNoticeList(params: NoticeListParams) {
  return request('/api/sys/notice/queryNoticeList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
