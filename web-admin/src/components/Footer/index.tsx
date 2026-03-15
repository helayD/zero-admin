import { DefaultFooter } from '@ant-design/pro-layout';

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear();

  return (
    <DefaultFooter
      copyright={`${currentYear} 九克城 · 企业级管理系统`}
      links={[]}
    />
  );
};

export default Footer;
