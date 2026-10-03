import { Component, type ErrorInfo, type ReactNode } from 'react';

import { reportEvent } from '../../lib/telemetry';
import { Alert } from '../ui/Alert';
import { Button } from '../ui/Button';

interface State {
  error: Error | null;
}

/**
 * Keeps the shell on screen when a page throws — most often a lazy page whose
 * chunk vanished because the site was redeployed while the tab was open.
 * Layout keys it by pathname, so moving to another page clears the error.
 */
export class RouteErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Page failed to render', error, info.componentStack);
    reportEvent({
      type: 'error',
      name: 'render',
      startTime: Date.now(),
      message: error.message,
      stack: info.componentStack ?? error.stack,
    });
  }

  render() {
    if (this.state.error) {
      return (
        <Alert tone="error" className="flex flex-wrap items-center justify-between gap-3">
          <span>
            Sorry, this page couldn&apos;t be loaded. The site may have just been updated.
          </span>
          <Button size="sm" variant="secondary" onClick={() => window.location.reload()}>
            Reload
          </Button>
        </Alert>
      );
    }
    return this.props.children;
  }
}
