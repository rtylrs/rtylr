import React, { useEffect, useState } from "react";
import type { LucideIcon } from "lucide-react";
import { Banknote, CreditCard, Smartphone } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { motion } from "motion/react";
import { toast } from "sonner";
import { createOneKitClient } from "@voxire/api";
import { OrderServiceClient } from "@voxire/voxie/pos/orders/v1/client";
import { OrderPaymentServiceClient } from "@voxire/voxie/pos/orders/payments/v1/client";
import POSLayout from "../../layouts/POSLayout";
import { useCartStore } from "../../shared/stores/cartStore";
import { useSettingsStore } from "../../shared/stores/settingsStore";
import { useOfflineQueueStore } from "../../shared/stores/offlineQueueStore";
import { useTerminalContextStore } from "../../shared/stores/terminalContextStore";
import {
  buildCashTender,
  buildCheckoutOrder,
  buildCreateOrderPayload,
  processCheckout,
  type CheckoutPaymentMethod,
} from "./checkoutWorkflow";
import { buildReceiptFromOrder } from "./receiptArtifact";
import { useReceiptsStore } from "../../shared/stores/receiptsStore";
import { useCashDrawerStore } from "../../shared/stores/cashDrawerStore";
import { printReceiptAndOpenDrawer, openDrawer } from "./paymentHardware";
import { resetAfterCheckout } from "./checkoutState";
import QrPayPanel from "./QrPayPanel";
import TipSelector from "../../shared/components/TipSelector";
import { earnPointsAfterPayment } from "../loyalty/loyaltyWorkflow";
import { resolveOpenSessionId } from "../close-session/closeSessionWorkflow";
import {
  buildOrderCompatPayload,
  buildResumeOrderPayload,
  createOrderCompat,
  requiresCompatOrder,
  persistResumedOrder,
} from "../orders/orderCompatWorkflow";

const orderClient = createOneKitClient(OrderServiceClient, "pos");
const orderPaymentClient = createOneKitClient(OrderPaymentServiceClient, "pos");

const fmt = (n: number): string =>
  new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(
    n,
  );

const Payment: React.FC = () => {
  const navigate = useNavigate();
  // Slice selectors keep the checkout screen from re-rendering on unrelated
  // cart writes (branchId, kitchen notes, persisted-hydration).
  const items = useCartStore((s) => s.items);
  const subtotalFn = useCartStore((s) => s.subtotal);
  const taxTotalFn = useCartStore((s) => s.taxTotal);
  const discountTotalFn = useCartStore((s) => s.discountTotal);
  const totalFn = useCartStore((s) => s.total);
  const storedTip = useCartStore((s) => s.tipAmount);
  const clearCart = useCartStore((s) => s.clearCart);
  const newOrder = useCartStore((s) => s.newOrder);
  const customerId = useCartStore((s) => s.customerId);
  const customerName = useCartStore((s) => s.customerName);
  const orderType = useCartStore((s) => s.orderType);
  const tableId = useCartStore((s) => s.tableId);
  const notes = useCartStore((s) => s.notes);
  const kitchenNotes = useCartStore((s) => s.kitchenNotes);
  const locationId = useTerminalContextStore((s) => s.locationId);
  // Stable local reference for the QR panel - regenerating it per render
  // would change the QR code under the customer's camera.
  const [localOrderRef] = useState(() => `local-${Date.now().toString(36)}`);
  const printer = useSettingsStore((s) => s.printer);
  const business = useSettingsStore((s) => s.business);
  const tipsEnabled = useSettingsStore((s) => s.features.tips);
  const tablesEnabled = useSettingsStore((s) => s.features.tables);
  const cashDrawerEnabled = useSettingsStore((s) => s.features.cashDrawer);
  const enqueueOffline = useOfflineQueueStore((s) => s.enqueue);
  const setServerOrderId = useCartStore((s) => s.setServerOrderId);
  const addReceipt = useReceiptsStore((s) => s.add);
  const recordCashSale = useCashDrawerStore((s) => s.recordCashSale);
  const [method, setMethod] = useState<CheckoutPaymentMethod>("card");
  const [processing, setProcessing] = useState(false);
  const [cashAmount, setCashAmount] = useState("");
  const [qrConfirmed, setQrConfirmed] = useState(false);
  const [tip, setTip] = useState(0);

  const subtotal = subtotalFn();
  const tax = taxTotalFn();
  const discount = discountTotalFn();
  const appliedTip = tipsEnabled ? tip : 0;
  const grandTotal = Math.max(0, totalFn() - storedTip + appliedTip);
  const cashTender = buildCashTender(cashAmount, grandTotal);
  const serverOrderId = useCartStore((s) => s.serverOrderId);

  const handleCharge = async (): Promise<void> => {
    setProcessing(true);
    try {
      const order = buildCheckoutOrder({
        id: localOrderRef,
        items,
        subtotal,
        tax,
        discount,
        total: grandTotal,
        tip: appliedTip,
        customerId,
        createdAt: new Date().toISOString(),
      });

      const sessionId = serverOrderId ? null : await resolveOpenSessionId();
      const payload = buildCreateOrderPayload({
        items,
        method,
        locationId,
        sessionId,
        tableId: tablesEnabled ? tableId : null,
        orderType,
        customerId,
        customerName,
        notes,
        kitchenNotes,
      });

      const result = await processCheckout({
        order,
        payload,
        method,
        existingOrderId: serverOrderId,
        persistExistingOrder: serverOrderId
          ? async () => {
              await persistResumedOrder(
                buildResumeOrderPayload({
                  orderId: serverOrderId,
                  items,
                  locationId,
                  tableId: tablesEnabled ? tableId : null,
                  orderType,
                  customerId,
                  customerName,
                  notes,
                  kitchenNotes,
                }),
              );
            }
          : undefined,
        createOrder: async (generatedPayload) => {
          if (requiresCompatOrder(items)) {
            return createOrderCompat(
              buildOrderCompatPayload({
                items,
                sessionId,
                locationId,
                tableId: tablesEnabled ? tableId : null,
                orderType,
                customerId,
                customerName,
                paymentMethod: method,
                notes,
                kitchenNotes,
              }),
            );
          }
          return orderClient.createOrder(generatedPayload);
        },
        chargeOrder: async (chargePayload) => {
          await orderPaymentClient.processPayment(chargePayload);
        },
        enqueueOffline: (offlinePayload) => {
          if (requiresCompatOrder(items)) {
            throw new Error("Combo orders require an online connection.");
          }
          enqueueOffline(offlinePayload);
        },
        recordCashSale,
        onOrderCreated: setServerOrderId,
      });

      const isCash = method === "cash";

      addReceipt(
        buildReceiptFromOrder({
          order,
          method,
          printedAt: new Date().toISOString(),
        }),
      );
      resetAfterCheckout({ newOrder, clearCart });
      if (result.status === "queued") {
        toast.warning(
          "Cash order saved offline. It will sync when connection returns.",
        );
      }

      // Earn loyalty points for attached customer (fire-and-forget, non-blocking)
      if (result.status === "online") {
        void earnPointsAfterPayment({
          customerId,
          orderId: result.orderId,
          totalAmount: grandTotal,
        })
          .then((earn) => {
            if (earn && earn.earned > 0) {
              toast.success(`+${earn.earned} loyalty points earned`);
            }
          })
          .catch(() => {
            // Points are non-critical; never surface an unhandled rejection
            // after a successful checkout.
          });
      }

      // Print receipt + open drawer via configured hardware
      const hwResult = await printReceiptAndOpenDrawer({
        printer,
        items,
        subtotal,
        taxAmount: tax,
        discountAmount: discount,
        tipAmount: appliedTip,
        total: grandTotal,
        paymentMethod: method,
        amountPaid: isCash ? cashTender.received || grandTotal : grandTotal,
        changeDue: isCash ? (cashTender.change ?? 0) : 0,
        orderNumber: order.id,
        businessName: business.name,
        footerMessage: undefined,
        kickDrawer: isCash && cashDrawerEnabled && printer.cashDrawerKick,
      });

      if (hwResult.kind === "print_failed") {
        toast.warning(`Receipt not printed: ${hwResult.error}`);
        // Still open drawer if printing failed but drawer was requested
        if (isCash && cashDrawerEnabled && printer.cashDrawerKick) {
          const dr = await openDrawer(printer);
          if (dr.kind === "drawer_failed") {
            toast.warning(`Cash drawer did not open: ${dr.error}`);
          }
        }
      } else if (hwResult.kind === "drawer_failed") {
        toast.warning(`Cash drawer did not open: ${hwResult.error}`);
      }

      navigate("/receipt");
    } catch (error) {
      toast.error(
        error instanceof Error
          ? error.message
          : "Payment failed. Please try again.",
      );
    } finally {
      setProcessing(false);
    }
  };

  useEffect(() => {
    if (items.length === 0) {
      navigate("/", { replace: true });
    }
  }, [items.length, navigate]);

  if (items.length === 0) return null;

  const methods: {
    id: CheckoutPaymentMethod;
    label: string;
    icon: LucideIcon;
  }[] = [
    { id: "card", label: "Card", icon: CreditCard },
    { id: "cash", label: "Cash", icon: Banknote },
    { id: "qr", label: "QR Pay", icon: Smartphone },
  ];

  return (
    <POSLayout>
      <div className="h-full flex items-center justify-center p-6">
        <div className="w-full max-w-md bg-pos-card border border-pos-border rounded-2xl p-6 space-y-5">
          <button
            className="text-sm text-pos-muted hover:text-pos-text transition-colors"
            onClick={() => navigate("/")}
          >
            ← Back to POS
          </button>
          <div className="flex items-center justify-between">
            <h1 className="text-xl font-bold text-pos-text">Payment</h1>
            <div className="flex gap-2">
              <button
                className="text-xs text-pos-muted hover:text-pos-text border border-pos-border rounded-lg px-3 py-1.5 transition-colors"
                onClick={() =>
                  navigate(
                    serverOrderId
                      ? `/split-bill/${serverOrderId}`
                      : "/split-bill/current",
                  )
                }
              >
                Split Payment
              </button>
              {serverOrderId && (
                <button
                  className="text-xs text-pos-muted hover:text-pos-text border border-pos-border rounded-lg px-3 py-1.5 transition-colors"
                  onClick={() => navigate(`/payment/${serverOrderId}`)}
                >
                  Multi-pay
                </button>
              )}
            </div>
          </div>

          {/* Summary */}
          <div className="space-y-2 text-sm">
            {items.map((item) => (
              <div
                key={item.id}
                className="flex justify-between text-pos-muted"
              >
                <span>
                  {item.quantity}× {item.name}
                </span>
                <span>{fmt(item.price * item.quantity)}</span>
              </div>
            ))}
            <div className="h-px bg-pos-border my-2" />
            <div className="flex justify-between text-pos-muted">
              <span>Subtotal</span>
              <span>{fmt(subtotal)}</span>
            </div>
            {discount > 0 && (
              <div className="flex justify-between text-pos-muted">
                <span>Discount</span>
                <span>-{fmt(discount)}</span>
              </div>
            )}
            <div className="flex justify-between text-pos-muted">
              <span>Tax</span>
              <span>{fmt(tax)}</span>
            </div>
            {appliedTip > 0 && (
              <div className="flex justify-between text-pos-muted">
                <span>Tip</span>
                <span>{fmt(appliedTip)}</span>
              </div>
            )}
            <div className="flex justify-between font-bold text-pos-text text-base pt-1">
              <span>Total</span>
              <span>{fmt(grandTotal)}</span>
            </div>
          </div>

          {tipsEnabled && (
            <TipSelector subtotal={subtotal} onTipChange={setTip} />
          )}

          {/* Method selector */}
          <div className="flex gap-2">
            {methods.map((m) => {
              const Icon = m.icon;
              return (
                <button
                  key={m.id}
                  className={`flex-1 inline-flex items-center justify-center gap-2 py-2.5 rounded-xl text-sm font-medium border transition-colors ${
                    method === m.id
                      ? "bg-pos-text text-pos-bg border-transparent"
                      : "border-pos-border text-pos-muted hover:border-white/20"
                  }`}
                  onClick={() => setMethod(m.id)}
                >
                  <Icon className="h-4 w-4" aria-hidden="true" />
                  {m.label}
                </button>
              );
            })}
          </div>

          {method === "card" && (
            <p className="text-sm text-pos-muted text-center">
              Present card to terminal or tap to pay.
            </p>
          )}

          {method === "cash" && (
            <div className="space-y-2">
              <label className="text-xs font-semibold uppercase tracking-wide text-pos-muted">
                Amount received
              </label>
              <input
                className="w-full bg-pos-input border border-pos-border rounded-xl px-4 py-3 text-pos-text text-sm outline-none focus:border-white/20 transition-colors"
                type="number"
                step="0.01"
                min={grandTotal}
                value={cashAmount}
                onChange={(e) => setCashAmount(e.target.value)}
                placeholder={fmt(grandTotal)}
              />
              {cashTender.change !== null && cashTender.change >= 0 && (
                <motion.div
                  className="text-center text-sm text-success font-medium"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                >
                  Change: {fmt(cashTender.change)}
                </motion.div>
              )}
            </div>
          )}

          {method === "qr" && (
            <QrPayPanel
              orderId={localOrderRef}
              amount={grandTotal}
              onConfirmed={() => setQrConfirmed(true)}
            />
          )}

          {method === "qr" && qrConfirmed && (
            <motion.div
              className="text-center text-sm text-emerald-400 font-medium"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
            >
              Payment confirmed - tap Charge to complete
            </motion.div>
          )}

          <button
            className="w-full py-3.5 rounded-xl bg-pos-text text-pos-bg font-bold text-sm hover:opacity-90 disabled:opacity-40 transition-opacity"
            onClick={handleCharge}
            disabled={
              processing ||
              (method === "cash" && !cashTender.canCharge) ||
              (method === "qr" && !qrConfirmed)
            }
          >
            {processing ? (
              <span className="inline-flex items-center gap-2">
                <span className="w-4 h-4 border-2 border-pos-bg border-t-transparent rounded-full animate-spin-slow" />
                Processing…
              </span>
            ) : (
              `Confirm Payment · ${fmt(grandTotal)}`
            )}
          </button>
        </div>
      </div>
    </POSLayout>
  );
};

export default Payment;
