import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import InboxScreen from '../features/inbox/Inbox';
import Tickets from '../features/tickets/Tickets';
import TicketDetail from '../features/tickets/TicketDetail';
import Customers from '../features/customers/Customers';
import KnowledgeBase from '../features/knowledge-base/KnowledgeBase';
import ArticleEditor from '../features/knowledge-base/ArticleEditor';
import CannedResponses from '../features/canned-responses/CannedResponses';
import Settings from '../features/settings/Settings';

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
      <Routes>
        <Route path="/" element={<InboxScreen />} />
        <Route path="/tickets" element={<Tickets />} />
        <Route path="/tickets/:id" element={<TicketDetail />} />
        <Route path="/customers" element={<Customers />} />
        <Route path="/knowledge" element={<KnowledgeBase />} />
        <Route path="/knowledge/new" element={<ArticleEditor />} />
        <Route path="/knowledge/:id/edit" element={<ArticleEditor />} />
        <Route path="/canned" element={<CannedResponses />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </DefaultLayout>
  );
};

export default AppRoutes;
