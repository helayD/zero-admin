export interface SystemOSSConfig {
  endpoint?: string;
  accessKeyId?: string;
  accessKeySecret?: string;
  accessKeySecretConfigured?: boolean;
  bucketName?: string;
  url?: string;
  maxSizeMb?: number;
}

export interface SystemSMSConfig {
  enabled?: boolean;
  provider?: string;
  endpoint?: string;
  accessKeyId?: string;
  accessKeySecret?: string;
  accessKeySecretConfigured?: boolean;
  signName?: string;
  templateCode?: string;
}

export interface SystemPushConfig {
  enabled?: boolean;
  provider?: string;
  endpoint?: string;
  appKey?: string;
  appSecret?: string;
  appSecretConfigured?: boolean;
}

export interface SystemConfigData {
  oss: SystemOSSConfig;
  sms: SystemSMSConfig;
  push: SystemPushConfig;
}

export interface QuerySystemConfigResponse {
  code: string;
  message: string;
  data: SystemConfigData;
}

export interface SaveSystemConfigResponse {
  code: string;
  message: string;
}
