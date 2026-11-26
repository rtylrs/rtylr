import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MenuLayoutCore from "./index";
import type { MenuCategoryData } from "../../components/MenuCategory";
import type { MenuLayoutConfig } from "../../types";

const categories: MenuCategoryData[] = [
  {
    id: "coffee",
    name: "Coffee",
    items: [{ id: "1", name: "Espresso", price: 3 }],
  },
];

const config: Partial<MenuLayoutConfig> = {
  pages: { gallery: true },
};

describe("MenuLayoutCore", () => {
  it("falls back to the menu when the active page is disabled by a config update", () => {
    vi.spyOn(window, "scrollTo").mockImplementation(() => undefined);
    const view = render(
      <MenuLayoutCore
        business={{ name: "Cafe" }}
        categories={categories}
        config={config}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Gallery" }));
    expect(screen.getByText("No photos yet.")).toBeTruthy();

    view.rerender(
      <MenuLayoutCore
        business={{ name: "Cafe" }}
        categories={categories}
        config={{ pages: { gallery: false } }}
      />,
    );

    expect(screen.getByRole("heading", { name: "Coffee" })).toBeTruthy();
    expect(screen.queryByText("No photos yet.")).toBeNull();
  });

  it("applies custom colors, scoped CSS, RTL direction, and keeps Powered by RTYLR visible", () => {
    const { container } = render(
      <MenuLayoutCore
        business={{ name: "Maison Cafe" }}
        categories={categories}
        activeLocale="ar"
        config={{
          colors: { primary: "#111827", accent: "#c8a96e", bg: "#101014" },
          customCss: ".rtylr-menu-root { letter-spacing: 0.02em; }",
        }}
      />,
    );

    const root = container.querySelector(".rtylr-menu-root") as HTMLElement;
    expect(root.getAttribute("dir")).toBe("rtl");
    expect(root.style.getPropertyValue("--menu-accent")).toBe("#c8a96e");
    expect(root.style.getPropertyValue("--menu-bg")).toBe("#101014");
    expect(screen.getByTestId("menu-custom-css").textContent).toContain("letter-spacing");
    expect(screen.getByText("Powered by RTYLR")).toBeTruthy();
  });

  it("switches currencies dynamically and triggers table actions", async () => {
    const onTableAction = vi.fn().mockResolvedValue(undefined);
    render(
      <MenuLayoutCore
        business={{ name: "Bistro" }}
        categories={categories}
        tableLabel="Table 4"
        onTableAction={onTableAction}
        banners={[{ id: "b1", title: "Happy Hour 20% Off" }]}
        stampCards={[
          { id: "s1", title: "Coffee Club", stampsRequired: 5, rewardLabel: "Free Espresso" },
        ]}
        currencies={[
          { code: "USD", symbol: "$", exchangeRate: 1, decimalPlaces: 2, isDefault: true },
          { code: "EUR", symbol: "€", exchangeRate: 0.9, decimalPlaces: 2 },
        ]}
      />,
    );

    expect(screen.getByText("Happy Hour 20% Off")).toBeTruthy();
    expect(screen.getByText(/Coffee Club/)).toBeTruthy();
    expect(screen.getByText("$3.00")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Language and currency" }));
    fireEvent.click(screen.getByRole("button", { name: "EUR (€)" }));
    expect(screen.getByText("€2.70")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Table service" }));
    fireEvent.click(screen.getByRole("button", { name: /Call Staff/ }));
    await waitFor(() => expect(onTableAction).toHaveBeenCalledWith("call_waiter"));
    expect(await screen.findByText(/Staff notified/)).toBeTruthy();
  });

  it("surfaces table action errors and closes sheets on Escape", async () => {
    const onTableAction = vi.fn().mockRejectedValue(new Error("network"));
    render(
      <MenuLayoutCore
        business={{ name: "Bistro" }}
        categories={categories}
        tableLabel="Table 4"
        onTableAction={onTableAction}
        stampCards={[
          { id: "s1", title: "Coffee Club", stampsRequired: 5, rewardLabel: "Free Espresso" },
        ]}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Table service" }));
    expect(screen.getByRole("dialog")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: /Call Staff/ }));
    expect(await screen.findByText(/Could not send request/)).toBeTruthy();

    fireEvent.keyDown(window, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("hides table pill when dine-in is off, avoids nested main in previewMode, and shows preview checkout notice", async () => {
    const { container } = render(
      <MenuLayoutCore
        business={{ name: "Bistro" }}
        categories={categories}
        tableLabel="Table 4"
        previewMode
        config={{
          features: {
            enableOrdering: true,
            fulfillmentModes: { dineIn: false, selfService: true, pickup: true, delivery: false },
          },
        }}
      />,
    );

    expect(screen.queryByText("Table 4")).toBeNull();
    expect(container.querySelector("main")).toBeNull();

    // Add item and open cart -> Place Order shows preview-only notice instead of silently closing.
    fireEvent.click(screen.getByRole("button", { name: /Add Espresso/i }));
    fireEvent.click(screen.getByRole("button", { name: /View Order/i }));
    fireEvent.click(screen.getByRole("button", { name: /Place Order/i }));
    expect(await screen.findByText(/Preview only/)).toBeTruthy();
  });

  it("validates and submits consenting guest details to CRM capture callback", async () => {
    const onCaptureCrmGuest = vi.fn().mockResolvedValue(undefined);
    render(
      <MenuLayoutCore
        business={{ name: "Bistro" }}
        categories={categories}
        onCaptureCrmGuest={onCaptureCrmGuest}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Stay in touch" }));
    // Empty submission shows validation alert and does not submit.
    fireEvent.click(screen.getByRole("button", { name: "Save my details" }));
    expect(screen.getByRole("alert").textContent).toContain("phone number or email");
    expect(onCaptureCrmGuest).not.toHaveBeenCalled();

    fireEvent.change(screen.getByLabelText("Full name"), { target: { value: "Nadia" } });
    fireEvent.change(screen.getByLabelText("Phone number"), { target: { value: "+15550002222" } });
    fireEvent.click(screen.getByRole("button", { name: "Save my details" }));

    await waitFor(() =>
      expect(onCaptureCrmGuest).toHaveBeenCalledWith({
        name: "Nadia",
        phone: "+15550002222",
        email: undefined,
        consent: true,
      }),
    );
    expect(await screen.findByText(/Your details have been saved/)).toBeTruthy();
  });
});
