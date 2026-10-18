import React from 'react';
import { Printer } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { PurchaseOrder } from '../../features/purchasing/usePurchaseOrders';
import { purchaseOrderLabel } from '../../features/purchasing/purchaseOrderStatusUtils';
import { toast } from 'sonner';
import { escapeHtml } from '../lib/escapeHtml';
import { formatCurrency, formatDate } from '../lib/format';

interface PrintPOButtonProps {
  order: PurchaseOrder;
  supplierName: string;
  productNameById: ReadonlyMap<string, string>;
  currency: string;
}

const PrintPOButton: React.FC<PrintPOButtonProps> = ({
  order,
  supplierName,
  productNameById,
  currency,
}) => {
  const { t } = useTranslation();

  const handlePrint = () => {
    const win = window.open('', '_blank', 'width=720,height=960');
    if (!win) {
      toast.error(t('erp.purchasing.print.popupBlocked'));
      return;
    }
    const money = (value: number | undefined) => escapeHtml(formatCurrency(value ?? 0, currency));
    // Every value below comes from user-entered data and is escaped: the
    // popup shares the portal's origin, so markup in a supplier or product
    // name would otherwise run as the signed-in user.
    const rows = (order.items ?? [])
      .map((item, index) => {
        const name =
          productNameById.get(item.productId ?? '') ??
          t('erp.purchasing.productFallback', { id: item.productId ?? '' });
        const quantity = item.quantityOrdered ?? 0;
        const unitCost = item.unitCost ?? 0;
        const lineTotal = item.totalCost ?? quantity * unitCost;
        return `<tr><td>${index + 1}</td><td>${escapeHtml(name)}</td><td class="num">${escapeHtml(quantity)}</td><td class="num">${money(unitCost)}</td><td class="num">${money(lineTotal)}</td></tr>`;
      })
      .join('');
    const notes = order.notes?.trim()
      ? `<p class="notes"><span class="label">${escapeHtml(t('erp.purchasing.print.notes'))}</span><br />${escapeHtml(order.notes)}</p>`
      : '';
    const html = `<!doctype html>
<html dir="${escapeHtml(document.documentElement.dir || 'ltr')}">
  <head>
    <meta charset="utf-8" />
    <title>${escapeHtml(t('erp.purchasing.print.title', { label: purchaseOrderLabel(order) }))}</title>
    <style>
      * { box-sizing: border-box; }
      body { font-family: Helvetica, Arial, sans-serif; padding: 32px; color: #111; }
      h1 { margin: 0 0 8px; font-size: 24px; }
      .row { display: flex; gap: 16px; margin: 20px 0; }
      .box { border: 1px solid #ddd; padding: 16px; border-radius: 8px; flex: 1; }
      .label { font-size: 10px; text-transform: uppercase; color: #888; letter-spacing: 0.05em; }
      .value { font-size: 14px; margin-top: 4px; font-weight: 600; }
      table { width: 100%; border-collapse: collapse; margin-top: 24px; }
      th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: start; font-size: 12px; }
      th { background: #f5f5f5; text-transform: uppercase; font-size: 10px; letter-spacing: 0.05em; }
      .num { text-align: end; }
      .totals { margin-top: 16px; margin-inline-start: auto; width: 260px; }
      .totals div { display: flex; justify-content: space-between; padding: 4px 0; font-size: 13px; }
      .totals .grand { font-size: 16px; font-weight: 700; border-top: 1px solid #ddd; padding-top: 8px; }
      .notes { margin-top: 24px; font-size: 12px; }
    </style>
  </head>
  <body>
    <h1>${escapeHtml(t('erp.purchasing.print.title', { label: purchaseOrderLabel(order) }))}</h1>
    <div class="row">
      <div class="box">
        <div class="label">${escapeHtml(t('erp.purchasing.columns.supplier'))}</div>
        <div class="value">${escapeHtml(supplierName)}</div>
      </div>
      <div class="box">
        <div class="label">${escapeHtml(t('erp.purchasing.print.orderDate'))}</div>
        <div class="value">${escapeHtml(order.createdAt ? formatDate(order.createdAt) : '')}</div>
      </div>
      <div class="box">
        <div class="label">${escapeHtml(t('erp.purchasing.columns.expected'))}</div>
        <div class="value">${escapeHtml(order.expectedDelivery ? formatDate(order.expectedDelivery) : '')}</div>
      </div>
    </div>
    <table>
      <thead>
        <tr>
          <th>#</th>
          <th>${escapeHtml(t('erpCommon.product'))}</th>
          <th class="num">${escapeHtml(t('erpCommon.quantity'))}</th>
          <th class="num">${escapeHtml(t('erp.purchasing.print.unitCost'))}</th>
          <th class="num">${escapeHtml(t('erpCommon.total'))}</th>
        </tr>
      </thead>
      <tbody>${rows}</tbody>
    </table>
    <div class="totals">
      <div><span>${escapeHtml(t('erp.purchasing.print.subtotal'))}</span><span>${money(order.subtotal)}</span></div>
      <div><span>${escapeHtml(t('erp.purchasing.print.tax'))}</span><span>${money(order.taxAmount)}</span></div>
      <div class="grand"><span>${escapeHtml(t('erpCommon.total'))}</span><span>${money(order.total)}</span></div>
    </div>
    ${notes}
    <script>window.onload = () => window.print();</script>
  </body>
</html>`;
    win.document.write(html);
    win.document.close();
  };

  return (
    <button
      type="button"
      onClick={handlePrint}
      className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium"
      style={{
        background: 'var(--hub-hover)',
        color: 'var(--hub-text-muted)',
        border: '1px solid var(--hub-border)',
      }}
    >
      <Printer size={11} /> {t('erp.purchasing.print.button')}
    </button>
  );
};

export default PrintPOButton;
