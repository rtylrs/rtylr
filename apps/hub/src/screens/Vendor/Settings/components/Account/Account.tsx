import React, { useState, useEffect } from 'react';
import { useUserStore } from '../../../../../stores/userStore';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Input,
  Label,
  Button,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../../../../components/ui';

interface LoginHistoryItem {
  email: string;
  method: string;
  ip: string;
  timeAgo: string;
}

const Account: React.FC = () => {
  const { user } = useUserStore();
  const [formData, setFormData] = useState({ firstName: '', lastName: '', email: '' });
  const [isEditing, setIsEditing] = useState(false);
  const [loginHistory] = useState<LoginHistoryItem[]>([
    { email: user?.email || '', method: 'Logged in via Google', ip: 'from 141.105.87.0', timeAgo: '3 days ago' },
    { email: user?.email || '', method: 'Logged in via Google', ip: 'from 141.105.87.0', timeAgo: '2 weeks ago' },
  ]);

  useEffect(() => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(' ')[0] || '',
        lastName: user.lastName || user.name?.split(' ').slice(1).join(' ') || '',
        email: user.email || '',
      });
    }
  }, [user]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    setIsEditing(true);
  };

  const handleSave = () => {
    console.log('Saving:', formData);
    setIsEditing(false);
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        firstName: user.firstName || user.name?.split(' ')[0] || '',
        lastName: user.lastName || user.name?.split(' ').slice(1).join(' ') || '',
        email: user.email || '',
      });
    }
    setIsEditing(false);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-primary">Account</h1>
        <p className="text-sm text-gray-400 mt-0.5">Manage your personal account information</p>
      </div>

      {/* Profile */}
      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Update your personal details</CardDescription>
        </CardHeader>
        <CardContent className="pt-0 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="firstName">First name</Label>
              <Input id="firstName" name="firstName" value={formData.firstName} onChange={handleInputChange} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="lastName">Last name</Label>
              <Input id="lastName" name="lastName" value={formData.lastName} onChange={handleInputChange} />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="email">Email</Label>
            <Input id="email" name="email" type="email" value={formData.email} onChange={handleInputChange} />
          </div>
          <div className="flex justify-end gap-2 pt-1">
            <Button variant="outline" onClick={handleCancel} disabled={!isEditing}>Cancel</Button>
            <Button onClick={handleSave} disabled={!isEditing}>Save changes</Button>
          </div>
        </CardContent>
      </Card>

      {/* Security */}
      <Card>
        <CardHeader>
          <CardTitle>Security</CardTitle>
          <CardDescription>Manage your account sessions</CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <p className="text-sm text-gray-500 mb-3">
            Invalidate all sessions for your account. This will log you out everywhere.
          </p>
          <Button variant="destructive" size="sm">Invalidate all sessions</Button>
        </CardContent>
      </Card>

      {/* Login History */}
      <Card>
        <CardHeader>
          <CardTitle>Login history</CardTitle>
          <CardDescription>Recent login activity for your account</CardDescription>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Method</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Time</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loginHistory.map((item, i) => (
                <TableRow key={i}>
                  <TableCell>{item.email}</TableCell>
                  <TableCell className="text-gray-500">{item.method}</TableCell>
                  <TableCell className="text-gray-500">{item.ip}</TableCell>
                  <TableCell className="text-gray-400">{item.timeAgo}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default Account;
