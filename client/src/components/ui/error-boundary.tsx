"use client";

import { Component, type ReactNode } from "react";

type Props = { fallback: (reset: () => void) => ReactNode; children: ReactNode };
type State = { failed: boolean };

// Catches a render error in its children, so one broken block shows its
// fallback instead of blanking the whole page. React only offers this as a
// class component. `reset` clears the error and renders the children again.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  componentDidCatch(error: unknown) {
    console.error(error);
  }

  reset = () => this.setState({ failed: false });

  render() {
    return this.state.failed ? this.props.fallback(this.reset) : this.props.children;
  }
}
