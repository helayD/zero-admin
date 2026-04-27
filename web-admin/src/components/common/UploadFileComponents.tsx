import { PlusOutlined } from '@ant-design/icons';
import { message, Modal, Upload } from 'antd';
import type { RcFile, UploadProps } from 'antd/es/upload';
import type { UploadFile } from 'antd/es/upload/interface';
import React, { useEffect, useMemo, useState } from 'react';

const defaultUploadApi = '/api/sys/upload';

const getBase64 = (file: RcFile): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = (error) => reject(error);
  });

const splitUrls = (value?: string) =>
  (value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);

const buildFileList = (value?: string): UploadFile[] =>
  splitUrls(value).map((url, index) => ({
    uid: `-${index + 1}`,
    name: url.substring(url.lastIndexOf('/') + 1) || `file-${index + 1}`,
    status: 'done',
    url,
  }));

const getFileUrl = (file: UploadFile) => {
  if (typeof file.response?.data === 'string') {
    return file.response.data;
  }
  return file.url;
};

// 文件上传
export interface UploadFileFormProps {
  // 上传成功后回调 url，兼容旧用法
  onSubmit?: (imageUrl: string) => void;
  // Form.Item 受控回调
  onChange?: (imageUrl: string) => void;
  // Form.Item 受控值，多个 URL 用逗号分割
  value?: string;
  // 需要显示多少个文件
  count?: number;
  // 上传接口
  backApi?: string;
  // 默认显示的文件
  defaultImageUrl?: string;
  accept?: string;
  listType?: UploadProps['listType'];
}

const UploadFileComponents: React.FC<UploadFileFormProps> = (props) => {
  const {
    accept = 'image/*',
    backApi = defaultUploadApi,
    count = 1,
    defaultImageUrl,
    listType = 'picture-card',
    onChange,
    onSubmit,
    value,
  } = props;

  const controlledValue = value ?? defaultImageUrl;
  const initialFileList = useMemo(() => buildFileList(controlledValue), [controlledValue]);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewImage, setPreviewImage] = useState('');
  const [previewTitle, setPreviewTitle] = useState('');
  const [fileList, setFileList] = useState<UploadFile[]>(initialFileList);

  useEffect(() => {
    setFileList(initialFileList);
  }, [initialFileList]);

  const emitChange = (nextFileList: UploadFile[]) => {
    const nextValue = nextFileList
      .filter((file) => file.status === 'done')
      .map(getFileUrl)
      .filter(Boolean)
      .join(',');

    onChange?.(nextValue);
    onSubmit?.(nextValue);
  };

  const handleCancel = () => setPreviewOpen(false);

  const handlePreview = async (file: UploadFile) => {
    if (!file.url && !file.preview && file.originFileObj) {
      file.preview = await getBase64(file.originFileObj as RcFile);
    }

    const url = file.url || (file.preview as string) || '';
    setPreviewImage(url);
    setPreviewOpen(true);
    setPreviewTitle(file.name || url.substring(url.lastIndexOf('/') + 1));
  };

  const handleChange: UploadProps['onChange'] = ({ file, fileList: nextFileList }) => {
    let uploadAccepted = true;
    if (file.status === 'done') {
      const { code, message: msg } = file.response || {};
      if (code === '000000') {
        message.success(msg || '上传成功');
      } else {
        uploadAccepted = false;
        message.error(msg || '上传失败');
      }
    }
    if (file.status === 'error') {
      message.error(`${file.name} 上传失败`);
    }

    const acceptedFileList = uploadAccepted
      ? nextFileList
      : nextFileList.filter((item) => item.uid !== file.uid);
    const normalizedFileList = acceptedFileList.slice(-count).map((item) => {
      const url = getFileUrl(item);
      if (!url) {
        return item;
      }
      return {
        ...item,
        name: item.name || url.substring(url.lastIndexOf('/') + 1),
        url,
      };
    });

    setFileList(normalizedFileList);
    emitChange(normalizedFileList);
  };

  const uploadButton = (
    <div>
      <PlusOutlined />
      <div style={{ marginTop: 8 }}>上传</div>
    </div>
  );

  return (
    <>
      <Upload
        accept={accept}
        action={backApi}
        listType={listType}
        fileList={fileList}
        onPreview={handlePreview}
        onChange={handleChange}
        headers={{
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        }}
      >
        {fileList.length >= count ? null : uploadButton}
      </Upload>
      <Modal open={previewOpen} title={previewTitle} footer={null} onCancel={handleCancel}>
        <img alt="preview" style={{ width: '100%' }} src={previewImage} />
      </Modal>
    </>
  );
};

export default UploadFileComponents;
