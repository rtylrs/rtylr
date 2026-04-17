import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, QrCode as QrCodeIcon, Copy, Eye } from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  Badge,
} from '../../shared/components/ui';

const QrCodes: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-primary">QR Codes</h1>
          <p className="text-sm text-gray-400 mt-0.5">Manage your table QR codes</p>
        </div>
        <Button size="sm" onClick={() => navigate('/app/business/qr-codes/create')}>
          <Plus size={14} />
          Generate QR Code
        </Button>
      </div>

      {/* Table */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle>All QR Codes</CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Table Number</TableHead>
                <TableHead>URL</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={5}>
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <div
                      className="h-16 w-16 rounded-2xl flex items-center justify-center mb-4"
                      style={{ background: 'var(--dash-hover)' }}
                    >
                      <QrCodeIcon size={28} className="text-gray-300" />
                    </div>
                    <p className="text-sm font-medium text-gray-400">No QR codes yet</p>
                    <p className="text-xs text-gray-300 mt-1 mb-4">
                      Generate QR codes for your tables to let customers order easily
                    </p>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => navigate('/app/business/qr-codes/create')}
                    >
                      <Plus size={13} />
                      Generate your first QR code
                    </Button>
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

export default QrCodes;
