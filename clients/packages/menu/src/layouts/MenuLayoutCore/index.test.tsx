import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MenuLayoutCore from "./index";
import type { MenuCategoryData } from "../../components/MenuCategory";
import { DEFAULT_LAYOUT_CONFIG, type MenuLayoutConfig } from "../../types";

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

  it("scrolls to the top of a clicked category section below the sticky header", () => {
    const scrollToSpy = vi.fn();
    const multiCategories: MenuCategoryData[] = [
      { id: "coffee", name: "Coffee", items: [{ id: "1", name: "Espresso", price: 3 }] },
      { id: "pastries", name: "Pastries", items: [{ id: "2", name: "Croissant", price: 4 }] },
    ];

    const { container } = render(
      <div data-testid="scroll-wrapper" style={{ overflowY: "auto", height: "600px" }}>
        <MenuLayoutCore
          business={{ name: "Cafe" }}
          categories={multiCategories}
          config={{ header: { ...DEFAULT_LAYOUT_CONFIG.header, sticky: true } }}
        />
      </div>,
    );

    const wrapper = screen.getByTestId("scroll-wrapper");
    Object.defineProperty(wrapper, "scrollHeight", { value: 2000, configurable: true });
    Object.defineProperty(wrapper, "clientHeight", { value: 600, configurable: true });
    Object.defineProperty(wrapper, "scrollTop", { value: 100, writable: true, configurable: true });
    wrapper.scrollTo = scrollToSpy;
    wrapper.getBoundingClientRect = () =>
      ({ top: 0, bottom: 600, left: 0, right: 400, width: 400, height: 600, x: 0, y: 0, toJSON: () => ({}) });

    const header = container.querySelector("header") as HTMLElement;
    header.getBoundingClientRect = () =>
      ({ top: 0, bottom: 120, left: 0, right: 400, width: 400, height: 120, x: 0, y: 0, toJSON: () => ({}) });

    const pastriesSection = container.querySelector("#category-pastries") as HTMLElement;
    pastriesSection.getBoundingClientRect = () =>
      ({ top: 500, bottom: 900, left: 0, right: 400, width: 400, height: 400, x: 0, y: 500, toJSON: () => ({}) });

    fireEvent.click(screen.getByRole("button", { name: "Pastries" }));

    // scrollTop(100) + (sectionTop(500) - containerTop(0)) - (headerHeight(120) + 12px padding) = 468
    expect(scrollToSpy).toHaveBeenCalledWith({ top: 468, behavior: "smooth" });
  });
});
