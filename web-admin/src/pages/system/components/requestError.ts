export const readErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};
