import * as Dialog from '@radix-ui/react-dialog';
import { useTranslation } from 'react-i18next';
import { FilterTestTarget, M3uSource } from '../types';
import { FilterTestPanelBody } from './FilterTestPanelBody';

interface FilterTestDrawerProps {
  target: FilterTestTarget | null;
  onOpenChange: (open: boolean) => void;
  sources: M3uSource[];
  modal?: boolean;
  withOverlay?: boolean;
}

// The Dialog.Root + .drawer-content shell around FilterTestPanelBody, mirroring
// MediaOccurrenceDrawer's split. Open exactly when target !== null; the title
// is target.label, which the owning component (CreateFilterDialog or
// FiltersSection) computes for its own context. See design.md.
export function FilterTestDrawer({ target, onOpenChange, sources, modal = true, withOverlay = true }: FilterTestDrawerProps) {
  const { t } = useTranslation('filters');

  return (
    <Dialog.Root open={target !== null} onOpenChange={onOpenChange} modal={modal}>
      <Dialog.Portal>
        {withOverlay && <Dialog.Overlay className="drawer-overlay" />}
        <Dialog.Content className="drawer-content">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1rem' }}>
            <Dialog.Title style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--primary-slate)', margin: 0 }}>
              {target?.label}
            </Dialog.Title>
            <Dialog.Close className="btn-secondary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem', borderRadius: 'var(--radius-sm)' }}>
              {t('dryRun.close')}
            </Dialog.Close>
          </div>

          <Dialog.Description style={{ display: 'none' }}>
            {t('dryRun.drawerDescription')}
          </Dialog.Description>

          {target && <FilterTestPanelBody target={target} sources={sources} />}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
