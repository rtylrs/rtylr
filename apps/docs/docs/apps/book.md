# Book (Scheduling)

The rtylr Book app handles appointment scheduling and resource booking. It is designed for service businesses such as salons, clinics, consultancies, and any business that takes appointments.

## Services

Define the services your business offers. Each service has a name, description, duration, price, and assigned staff members. Services can be grouped into categories (e.g., Haircuts, Coloring, Treatments). Buffer time can be added before or after a service to allow for setup and cleanup.

## Resources

Resources represent anything that needs to be reserved alongside a service: rooms, equipment, chairs, or stations. When a resource is required by a service, the system checks resource availability in addition to staff availability before allowing a booking.

## Appointments

Appointments are the core entity in Book. Each appointment includes the customer, service, assigned staff member, date, time, duration, and status. Statuses include `pending`, `confirmed`, `checked_in`, `in_progress`, `completed`, `cancelled`, and `no_show`. Appointments can be created by staff in the app or by customers through the online booking page.

## Calendar View

The calendar displays appointments in day, week, or month views. It can be filtered by staff member or resource. Drag-and-drop rescheduling is supported. Color coding distinguishes services and statuses at a glance.

## Waitlist

When all slots are booked, customers can join a waitlist. If a cancellation opens a slot, waitlisted customers are notified automatically. The waitlist is managed per service and date range.

## Customers

Book shares customer profiles with CRM. Appointment history, preferences, and notes are visible on the customer profile. Returning customers can be identified quickly and their past services referenced during booking.

## Dashboard

The Book dashboard shows today's schedule, upcoming appointments, recent cancellations, and key metrics (bookings this week, revenue, no-show rate). It provides a quick overview for managers and front desk staff.

## Settings

Configure scheduling rules and business hours:

- **Business hours** — Set operating hours per day of the week, with support for breaks
- **Booking window** — Define how far in advance customers can book (e.g., up to 30 days) and the minimum notice period (e.g., 2 hours before)
- **Cancellation policy** — Set the cancellation window and any fees
- **Reminders** — Configure automated appointment reminders via email or SMS
- **Online booking** — Enable or disable the public booking page and customize its appearance
