export const book = {
  title: 'Book (Scheduling)',
  body: `
<p>The rtylr Book app handles appointment scheduling and resource booking. It is designed for service businesses such as salons, clinics, consultancies, and any business that takes appointments.</p>

<h2 id="services">Services</h2>
<p>Define the services your business offers. Each service has a name, description, duration, price, and assigned staff members. Services can be grouped into categories (e.g., Haircuts, Coloring, Treatments). Buffer time can be added before or after a service to allow for setup and cleanup.</p>

<h2 id="resources">Resources</h2>
<p>Resources represent anything that needs to be reserved alongside a service: rooms, equipment, chairs, or stations. When a resource is required by a service, the system checks resource availability in addition to staff availability before allowing a booking.</p>

<h2 id="appointments">Appointments</h2>
<p>Appointments are the core entity in Book. Each appointment includes the customer, service, assigned staff member, date, time, duration, and status. Statuses include <code>pending</code>, <code>confirmed</code>, <code>checked_in</code>, <code>in_progress</code>, <code>completed</code>, <code>cancelled</code>, and <code>no_show</code>. Appointments can be created by staff in the app or by customers through the online booking page.</p>

<h2 id="calendar">Calendar View</h2>
<p>The calendar displays appointments in day, week, or month views. It can be filtered by staff member or resource. Drag-and-drop rescheduling is supported. Color coding distinguishes services and statuses at a glance.</p>

<h2 id="waitlist">Waitlist</h2>
<p>When all slots are booked, customers can join a waitlist. If a cancellation opens a slot, waitlisted customers are notified automatically. The waitlist is managed per service and date range.</p>

<h2 id="customers">Customers</h2>
<p>Book shares customer profiles with CRM. Appointment history, preferences, and notes are visible on the customer profile. Returning customers can be identified quickly and their past services referenced during booking.</p>

<h2 id="dashboard">Dashboard</h2>
<p>The Book dashboard shows today's schedule, upcoming appointments, recent cancellations, and key metrics (bookings this week, revenue, no-show rate). It provides a quick overview for managers and front desk staff.</p>

<h2 id="settings">Settings</h2>
<p>Configure scheduling rules and business hours:</p>
<ul>
  <li><strong>Business hours</strong> - Set operating hours per day of the week, with support for breaks</li>
  <li><strong>Booking window</strong> - Define how far in advance customers can book (e.g., up to 30 days) and the minimum notice period (e.g., 2 hours before)</li>
  <li><strong>Cancellation policy</strong> - Set the cancellation window and any fees</li>
  <li><strong>Reminders</strong> - Configure automated appointment reminders via email or SMS</li>
  <li><strong>Online booking</strong> - Enable or disable the public booking page and customize its appearance</li>
</ul>
  `.trim(),
};
