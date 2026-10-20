import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate } from '@tanstack/react-router';
import { ChevronLeft, ChevronRight, Mail, Phone, X } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { dateOf, formatRelative, leadTitle, personName, telHref } from './data/model';
import { useLeadFilters } from './data/leadFilters';
import { useLead, useLeadList } from './data/queries';
import { DueDate, EmptyBlock, LeadStatusBadge, Money } from './components/common';
import { LogActivityPanel, Timeline } from './components/Activity';

/**
 * Works through the leads list one lead at a time: call, log the outcome, and
 * move on. The queue is the list as it was loaded, so leads whose status
 * changes mid-session stay in place instead of reshuffling the order.
 */
const CallSession: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const filters = useLeadFilters((state) => state.filters);
  const list = useLeadList(filters);
  const [queue, setQueue] = useState<string[]>([]);
  const [index, setIndex] = useState(0);
  const [called, setCalled] = useState(0);

  useEffect(() => {
    const ids =
      list.data?.pages.flatMap((page) => page.leads ?? []).map((lead) => lead.id ?? '') ?? [];
    setQueue((current) => {
      const seen = new Set(current);
      const added = ids.filter((id) => id && !seen.has(id));
      return added.length > 0 ? [...current, ...added] : current;
    });
  }, [list.data]);

  const total = list.data?.pages[0]?.total ?? 0;
  const currentId = queue[index];
  const lead = useLead(currentId);
  const nearEnd = index >= queue.length - 3;
  useEffect(() => {
    if (nearEnd && list.hasNextPage && !list.isFetchingNextPage) void list.fetchNextPage();
  }, [nearEnd, list]);

  const next = (): void => setIndex((value) => Math.min(value + 1, Math.max(queue.length - 1, 0)));
  const previous = (): void => setIndex((value) => Math.max(value - 1, 0));
  const finished = queue.length > 0 && index >= queue.length - 1 && !list.hasNextPage;
  const format = (value: number) => new Intl.NumberFormat(i18n.language).format(value);
  const data = lead.data;
  const phone = telHref(data?.phone);
  const subtitle = useMemo(
    () =>
      data
        ? [data.companyName ? personName(data) : '', data.jobTitle].filter(Boolean).join(' · ')
        : '',
    [data]
  );

  if (list.isPending) {
    return (
      <p className="sales-empty" role="status">
        {t('crmSales.common.loading')}
      </p>
    );
  }
  if (queue.length === 0) {
    return (
      <div className="sales-page sales-call">
        <EmptyBlock
          icon={Phone}
          title={t('crmSales.call.emptyTitle')}
          description={t('crmSales.call.emptyHint')}
        >
          <Button onClick={() => navigate({ to: '/crm/sales/leads' })}>
            {t('crmSales.call.backToLeads')}
          </Button>
        </EmptyBlock>
      </div>
    );
  }

  return (
    <div className="sales-page sales-call">
      <HelmetContainer title={`${t('crmSales.call.title')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.call.title')}</h1>
          <p>
            {t('crmSales.call.position', {
              position: format(index + 1),
              total: format(total || queue.length),
            })}
            {called > 0 && ` · ${t('crmSales.call.loggedCount', { count: called })}`}
          </p>
        </div>
        <div className="sales-actions">
          <Button
            variant="outline"
            onClick={previous}
            disabled={index === 0}
            aria-label={t('crmSales.call.previous')}
          >
            <ChevronLeft size={14} />
          </Button>
          <Button variant="outline" onClick={next} disabled={finished}>
            {t('crmSales.call.skip')}
            <ChevronRight size={14} />
          </Button>
          <Button variant="ghost" onClick={() => navigate({ to: '/crm/sales/leads' })}>
            <X size={14} />
            {t('crmSales.call.exit')}
          </Button>
        </div>
      </header>

      {!data ? (
        <p className="sales-empty" role="status">
          {lead.isError ? t('crmSales.leadDetail.loadFailed') : t('crmSales.common.loading')}
        </p>
      ) : (
        <>
          <section className="sales-panel">
            <div className="sales-panel-head">
              <div style={{ minWidth: 0 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                  <Link
                    className="sales-link"
                    to={`/crm/sales/leads/${data.id}`}
                    style={{ fontSize: 18, fontWeight: 700 }}
                  >
                    {leadTitle(data)}
                  </Link>
                  <LeadStatusBadge status={data.status} />
                </div>
                {subtitle && <div className="sales-cell-sub">{subtitle}</div>}
              </div>
              <Money value={data.value} billing={data.billing} empty="" />
            </div>
            {phone ? (
              <a className="sales-call-phone" href={phone}>
                <Phone size={20} aria-hidden="true" />
                <bdi dir="ltr">{data.phone}</bdi>
              </a>
            ) : (
              <p className="sales-muted" style={{ margin: 0 }}>
                {t('crmSales.call.noPhone')}
              </p>
            )}
            {data.email && (
              <a className="sales-contact-line" href={`mailto:${data.email}`}>
                <Mail size={14} aria-hidden="true" />
                <bdi dir="ltr">{data.email}</bdi>
              </a>
            )}
            <dl className="sales-fields">
              <dt>{t('crmSales.fields.lastActivity')}</dt>
              <dd>
                {data.lastActivityAt
                  ? formatRelative(dateOf(data.lastActivityAt), i18n.language)
                  : t('crmSales.leads.neverWorked')}
              </dd>
              <dt>{t('crmSales.fields.attempts')}</dt>
              <dd>{data.attemptCount ?? 0}</dd>
              <dt>{t('crmSales.fields.followUp')}</dt>
              <dd>
                <DueDate date={data.nextFollowUp} />
              </dd>
              {data.notes && (
                <>
                  <dt>{t('crmSales.fields.notes')}</dt>
                  <dd style={{ whiteSpace: 'pre-wrap' }}>{data.notes}</dd>
                </>
              )}
            </dl>
          </section>
          <LogActivityPanel
            key={data.id}
            subject={{ leadId: data.id }}
            compact
            onLogged={(outcome) => {
              if (outcome) setCalled((value) => value + 1);
              if (!finished) next();
            }}
          />
          <Timeline subject={{ leadId: data.id }} />
        </>
      )}
      {finished && (
        <div className="sales-banner" role="status">
          {t('crmSales.call.done')}
        </div>
      )}
    </div>
  );
};

export default CallSession;
