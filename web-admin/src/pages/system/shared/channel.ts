export const SYSTEM_CHANNEL_OPTIONS = [
  { label: 'App', value: 'app' },
  { label: 'PC', value: 'pc' },
  { label: 'H5', value: 'h5' },
  { label: '小程序', value: 'mini_program' },
  { label: '门店 POS', value: 'pos' },
];

export const SYSTEM_CHANNEL_VALUE_ENUM = {
  app: { text: 'App' },
  pc: { text: 'PC' },
  h5: { text: 'H5' },
  mini_program: { text: '小程序' },
  pos: { text: '门店 POS' },
  unknown: { text: '未知渠道' },
};

const SYSTEM_CHANNEL_LABEL_MAP: Record<string, string> = {
  app: 'App',
  pc: 'PC',
  h5: 'H5',
  mini_program: '小程序',
  pos: '门店 POS',
  unknown: '未知渠道',
};

export const normalizeSystemChannel = (channel?: string) => {
  const trimmed = (channel || '').trim();
  if (!trimmed) {
    return '';
  }
  if (trimmed === 'mini-program') {
    return 'mini_program';
  }
  if (trimmed in SYSTEM_CHANNEL_LABEL_MAP) {
    return trimmed;
  }
  return trimmed;
};

export const formatSystemChannelLabel = (channel?: string) => {
  const normalized = normalizeSystemChannel(channel);
  return SYSTEM_CHANNEL_LABEL_MAP[normalized] || normalized || '-';
};
