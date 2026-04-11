import React from 'react';
import { CreditCard, Download, Zap, CheckCircle } from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Badge,
  Button,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  Separator,
} from '../../shared/components/ui';

const planFeatures = ['Up to 3 menus', 'QR code generation', 'Basic analytics', 'Email support'];

const Billing: React.FC = () => {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-xl font-bold text-primary">Billing</h1>
        <p className="text-sm text-gray-400 mt-0.5">Manage your subscription and invoices</p>
      </div>

      {/* Current Plan */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>Starter Plan</CardTitle>
              <CardDescription className="mt-1">Your current subscription plan</CardDescription>
            </div>
            <Badge variant="success">Active</Badge>
          </div>
        </CardHeader>
        <CardContent className="pt-0 space-y-4">
          <div className="flex items-baseline gap-1">
            <span className="text-3xl font-bold text-primary">$0</span>
            <span className="text-sm text-gray-400">/month</span>
          </div>
          <Separator />
          <div className="grid grid-cols-2 gap-2">
            {planFeatures.map((feature) => (
              <div
                key={feature}
                className="flex items-center gap-2 text-sm"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                <CheckCircle size={14} className="text-emerald-500 shrink-0" />
                {feature}
              </div>
            ))}
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-gray-400">Next renewal</p>
              <p className="text-sm font-medium text-primary mt-0.5">-</p>
            </div>
            <Button variant="accent" size="sm">
              <Zap size={14} />
              Upgrade to Pro
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Payment Method */}
      <Card>
        <CardHeader>
          <CardTitle>Payment Method</CardTitle>
          <CardDescription>Your saved payment methods</CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <div
            className="flex flex-col items-center justify-center py-8 text-center rounded-lg"
            style={{ border: '1px dashed var(--hub-border)' }}
          >
            <CreditCard size={28} className="text-gray-300 mb-3" />
            <p className="text-sm font-medium text-gray-400">No payment method added</p>
            <p className="text-xs text-gray-300 mt-1 mb-4">Add a card to upgrade your plan</p>
            <Button variant="outline" size="sm">
              <CreditCard size={13} />
              Add Payment Method
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Invoices */}
      <Card>
        <CardHeader>
          <CardTitle>Invoices</CardTitle>
          <CardDescription>Your billing history</CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Download</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={5}>
                  <div className="flex flex-col items-center justify-center py-8 text-center">
                    <Download size={28} className="text-gray-200 mb-3" />
                    <p className="text-sm font-medium text-gray-400">No invoices yet</p>
                    <p className="text-xs text-gray-300 mt-1">Your invoices will appear here</p>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default Billing;
