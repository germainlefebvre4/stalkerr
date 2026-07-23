import { useState, useCallback } from 'react';

export interface ToastNotification {
  message: string;
  type: 'success' | 'error';
}

export function useToast() {
  const [notification, setNotification] = useState<ToastNotification | null>(null);

  const showToast = useCallback((message: string, type: 'success' | 'error' = 'success') => {
    setNotification({ message, type });
    setTimeout(() => {
      setNotification(null);
    }, 4000);
  }, []);

  return {
    notification,
    showToast
  };
}
