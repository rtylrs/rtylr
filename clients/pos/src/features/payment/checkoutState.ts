interface ResetAfterCheckoutParams {
  newOrder: () => void;
  clearCart: () => void;
}

export function resetAfterCheckout({ newOrder }: ResetAfterCheckoutParams): void {
  newOrder();
}
