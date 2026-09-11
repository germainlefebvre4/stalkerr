import { useState, useCallback, useEffect } from 'react';

const REDUCE_MOTION_STORAGE_KEY = 'stalkeer_reduce_motion';

function applyReduceMotion(enabled: boolean) {
  if (enabled) {
    document.documentElement.dataset.reduceMotion = 'true';
  } else {
    delete document.documentElement.dataset.reduceMotion;
  }
}

function readStoredReduceMotion(): boolean {
  return localStorage.getItem(REDUCE_MOTION_STORAGE_KEY) === 'true';
}

export function useReduceMotion() {
  const [reduceMotion, setReduceMotionState] = useState<boolean>(readStoredReduceMotion);

  useEffect(() => {
    applyReduceMotion(reduceMotion);
  }, [reduceMotion]);

  const setReduceMotion = useCallback((next: boolean) => {
    localStorage.setItem(REDUCE_MOTION_STORAGE_KEY, next ? 'true' : 'false');
    setReduceMotionState(next);
  }, []);

  return { reduceMotion, setReduceMotion };
}
