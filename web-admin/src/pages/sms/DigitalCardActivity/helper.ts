import moment from 'moment';
import type {
  DrawActivityFormValues,
  DrawActivityListItem,
  PreviewDrawActivityPublishReadinessData,
} from './data.d';

const formatDateTime = (value?: moment.MomentInput) => {
  if (!value) {
    return '';
  }
  return moment(value).format('YYYY-MM-DD HH:mm:ss');
};

export const serializeDrawActivityPayload = (values: DrawActivityFormValues) => {
  const startValue = Array.isArray(values.activeTime) ? values.activeTime[0] : values.startTime;
  const endValue = Array.isArray(values.activeTime) ? values.activeTime[1] : values.endTime;

  return {
    id: values.id,
    activityCode: values.activityCode,
    name: values.name,
    ruleSummary: values.ruleSummary || '',
    startTime: formatDateTime(startValue),
    endTime: formatDateTime(endValue),
    realNameRequired: values.realNameRequired ?? 0,
    participantConditionSummary: values.participantConditionSummary || '',
    consumeRuleSummary: values.consumeRuleSummary || '',
    probabilityRule: values.probabilityRule || '',
    complianceRuleSummary: values.complianceRuleSummary || '',
    circulationLimitSummary: values.circulationLimitSummary || '',
    approvalRecordRef: values.approvalRecordRef || '',
    copyrightStatus: values.copyrightStatus ?? 0,
    contentAuditStatus: values.contentAuditStatus ?? 0,
    auditStatus: values.auditStatus ?? 0,
    status: values.status ?? 0,
    isEnabled: values.isEnabled ?? 1,
    homeEntry: {
      showOnHome: values.homeEntry?.showOnHome ?? values.showOnHome ?? 0,
      homeEntryTitle: values.homeEntry?.homeEntryTitle || values.homeEntryTitle || '',
      homeEntrySubtitle: values.homeEntry?.homeEntrySubtitle || '',
      homeEntryImage: values.homeEntry?.homeEntryImage || '',
      homeEntrySort: values.homeEntry?.homeEntrySort ?? 0,
      landingTargetType: values.homeEntry?.landingTargetType || '',
      landingTargetValue: values.homeEntry?.landingTargetValue || '',
      isEnabled: values.homeEntry?.isEnabled ?? 1,
    },
    templates: (values.templates || []).map((item) => ({
      id: item.id,
      templateName: item.templateName,
      templateCode: item.templateCode,
      cardFaceImage: item.cardFaceImage || '',
      copyrightOwner: item.copyrightOwner || '',
      copyrightProofSummary: item.copyrightProofSummary || '',
      rarity: item.rarity || '',
      issueLimit: Number(item.issueLimit || 0),
      displayCopy: item.displayCopy || '',
      circulationLimitSummary: item.circulationLimitSummary || '',
      displayStatus: item.displayStatus ?? 1,
      contentAuditStatus: item.contentAuditStatus ?? 0,
      providerCode: item.providerCode || '',
      credentialRef: item.credentialRef || '',
      status: item.status ?? 0,
      auditStatus: item.auditStatus ?? 0,
    })),
    pools: (values.pools || []).map((item) => ({
      id: item.id,
      poolName: item.poolName,
      poolCode: item.poolCode,
      probabilityRule: item.probabilityRule || '',
      sort: item.sort ?? 0,
      status: item.status ?? 0,
      templates: (item.templates || []).map((mapping) => ({
        id: mapping.id,
        templateId: mapping.templateId,
        templateCode: mapping.templateCode || '',
        templateName: mapping.templateName || '',
        rarity: mapping.rarity || '',
        probability: Number(mapping.probability || 0),
        saleLimit: Number(mapping.saleLimit || 0),
        remainingLimit: Number(mapping.remainingLimit || 0),
        configLimit: Number(mapping.configLimit || 0),
      })),
    })),
  };
};

export const buildPreviewMessages = (preview?: PreviewDrawActivityPublishReadinessData) => {
  if (!preview) {
    return [];
  }
  if (!Array.isArray(preview.items) || preview.items.length === 0) {
    return [preview.summary || '发布预检通过'];
  }
  return preview.items.map((item) => `${item.blocking ? '阻断' : '待处理'}: ${item.message}`);
};

export const toFormInitialValues = (item?: DrawActivityListItem) => {
  if (!item) {
    return {
      status: 0,
      auditStatus: 0,
      isEnabled: 1,
      realNameRequired: 0,
      copyrightStatus: 0,
      contentAuditStatus: 0,
      activeTime: [],
      homeEntry: {
        showOnHome: 0,
        homeEntrySort: 0,
        isEnabled: 1,
      },
      templates: [],
      pools: [],
    };
  }
  return {
    ...item,
    activeTime:
      item.startTime && item.endTime ? [moment(item.startTime), moment(item.endTime)] : [],
    homeEntry: {
      showOnHome: item.homeEntry?.showOnHome ?? item.showOnHome ?? 0,
      homeEntryTitle: item.homeEntry?.homeEntryTitle || item.homeEntryTitle || '',
      homeEntrySubtitle: item.homeEntry?.homeEntrySubtitle || '',
      homeEntryImage: item.homeEntry?.homeEntryImage || '',
      homeEntrySort: item.homeEntry?.homeEntrySort ?? 0,
      landingTargetType: item.homeEntry?.landingTargetType || '',
      landingTargetValue: item.homeEntry?.landingTargetValue || '',
      isEnabled: item.homeEntry?.isEnabled ?? 1,
    },
    templates: item.templates || [],
    pools: item.pools || [],
  };
};
