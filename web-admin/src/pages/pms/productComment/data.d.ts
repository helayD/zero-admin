// 商品评价数据模型（Story 8-2）

export interface CommentListItem {
  id: string; // 评价ID
  productId: number; // 商品ID
  memberNickName: string; // 评价者昵称
  memberId: number; // 会员ID
  memberIcon: string; // 评价者头像
  star: number; // 评分 0-5
  content: string; // 评价内容
  pics: string; // 图片地址，逗号分隔
  productAttribute: string; // 购买商品属性
  showStatus: number; // 审核状态：0=屏蔽，1=通过
  replayCount: number; // 回复数量
  memberIp: string; // 评价IP
  createTime: string; // 评价时间
  updateBy?: string; // 处理人
}

export interface CommentReplayItem {
  id: string; // 回复ID
  commentId: string; // 关联评价ID（MongoDB ObjectID）
  type: number; // 评论人员类型：0->会员；1->管理员
  memberNickName: string; // 评论人员昵称
  memberIcon: string; // 评论人员头像
  content: string; // 内容
  createTime: string; // 回复时间
}

export interface CommentDetailData extends CommentListItem {
  replays?: CommentReplayItem[];
}

export interface CommentListData {
  list: CommentListItem[];
  pagination: {
    total: number;
    pageSize: number;
    current: number;
  };
}

export interface CommentListParams {
  productId?: number; // 商品ID
  showStatus?: number; // 审核状态：0=屏蔽，1=通过
  startTime?: string; // 开始时间
  endTime?: string; // 结束时间
  pageSize?: number;
  current?: number;
}

export interface CommentDetailParams {
  id: string; // 评价ID
}

export interface UpdateCommentParams {
  id: string; // 评价ID
  showStatus: number; // 审核状态：0=屏蔽，1=通过
  updateBy: string; // 处理人
}

export interface CommentDetailResponse {
  code: number;
  message: string;
  data: CommentDetailData;
  replays?: CommentReplayItem[];
}
