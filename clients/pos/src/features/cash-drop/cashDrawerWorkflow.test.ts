import { describe, expect, it, vi } from "vitest";
import {
  processCashDrawerAction,
  recordCashDrawerAction,
} from "./cashDrawerWorkflow";
import { NoOpenRegisterSessionError } from "../close-session/closeSessionWorkflow";
import type { CashMovementPayload } from "../../shared/stores/cashMovementQueueStore";

const actions = () => ({
  openDrawer: vi.fn(),
  cashDrop: vi.fn(),
  paidOut: vi.fn(),
  paidIn: vi.fn(),
});

const context = {
  branchId: 1,
  terminalId: 2,
  employeeId: 3,
  resolveSessionId: vi.fn(async () => 10),
};

describe("recordCashDrawerAction", () => {
  it("opens the drawer for a valid float", () => {
    const drawer = actions();

    const result = recordCashDrawerAction({
      action: "float",
      amountInput: "125.50",
      reason: "",
      drawer,
    });

    expect(result).toEqual({
      amount: 125.5,
      message: "Opening float set to $125.50",
    });
    expect(drawer.openDrawer).toHaveBeenCalledWith(125.5);
  });

  it("rejects non-finite amounts before mutating drawer state", () => {
    const drawer = actions();

    expect(() =>
      recordCashDrawerAction({
        action: "paid_in",
        amountInput: "abc",
        reason: "Correction",
        drawer,
      }),
    ).toThrow("Enter a valid amount greater than zero.");
    expect(drawer.paidIn).not.toHaveBeenCalled();
  });

  it("requires a reason for manual cash movements", () => {
    const drawer = actions();

    expect(() =>
      recordCashDrawerAction({
        action: "drop",
        amountInput: "20",
        reason: "   ",
        drawer,
      }),
    ).toThrow("Enter a reason for this cash movement.");
    expect(drawer.cashDrop).not.toHaveBeenCalled();
  });

  it("records a cash drop with trimmed reason", () => {
    const drawer = actions();

    const result = recordCashDrawerAction({
      action: "drop",
      amountInput: "20",
      reason: " Safe deposit ",
      drawer,
    });

    expect(result).toEqual({ amount: 20, message: "Cash drop recorded" });
    expect(drawer.cashDrop).toHaveBeenCalledWith(20, "Safe deposit");
  });
});

describe("processCashDrawerAction", () => {
  it("syncs manual cash movements after recording them locally", async () => {
    const drawer = actions();
    const syncMovement = vi.fn(async () => undefined);
    const enqueueMovement = vi.fn();

    const resolveSessionId = vi.fn(async () => 10);

    const result = await processCashDrawerAction({
      action: "paid_out",
      amountInput: "15",
      reason: "Petty cash",
      drawer,
      context: { branchId: 1, terminalId: 2, employeeId: 3, resolveSessionId },
      syncMovement,
      enqueueMovement,
    });

    expect(result).toEqual({ amount: 15, message: "Paid out", queued: false });
    expect(drawer.paidOut).toHaveBeenCalledWith(15, "Petty cash");

    const expectedPayload: CashMovementPayload = {
      branch_id: 1,
      terminal_id: 2,
      session_id: 10,
      employee_id: 3,
      amount: 15,
      reason: "Petty cash",
    };
    expect(syncMovement).toHaveBeenCalledWith("paid_out", expectedPayload);
    expect(enqueueMovement).not.toHaveBeenCalled();
  });

  it("enqueues manual cash movements when backend sync fails", async () => {
    const drawer = actions();
    const enqueueMovement = vi.fn();
    const resolveSessionId = vi.fn(async () => 10);

    const result = await processCashDrawerAction({
      action: "drop",
      amountInput: "20",
      reason: "Safe deposit",
      drawer,
      context: { branchId: 1, terminalId: 2, employeeId: 3, resolveSessionId },
      syncMovement: vi.fn(async () => {
        throw new Error("offline");
      }),
      enqueueMovement,
    });

    expect(result).toEqual({
      amount: 20,
      message: "Cash drop recorded",
      queued: true,
    });
    expect(drawer.cashDrop).toHaveBeenCalledWith(20, "Safe deposit");
    expect(enqueueMovement).toHaveBeenCalledWith({
      action: "drop",
      payload: expect.objectContaining({
        branch_id: 1,
        amount: 20,
        session_id: 0,
      }),
    });
  });

  it("does not mutate or enqueue when no register session is open", async () => {
    const drawer = actions();
    const enqueueMovement = vi.fn();

    await expect(
      processCashDrawerAction({
        action: "drop",
        amountInput: "20",
        reason: "Safe deposit",
        drawer,
        context: {
          branchId: 1,
          terminalId: 2,
          employeeId: 3,
          resolveSessionId: async () => {
            throw new NoOpenRegisterSessionError();
          },
        },
        syncMovement: vi.fn(),
        enqueueMovement,
      }),
    ).rejects.toThrow("No open register session found.");

    expect(drawer.cashDrop).not.toHaveBeenCalled();
    expect(enqueueMovement).not.toHaveBeenCalled();
  });

  it("rolls back the local drawer when the server definitively rejects a movement", async () => {
    const drawer = actions();
    const enqueueMovement = vi.fn();
    const rejection = Object.assign(new Error("permission denied"), {
      statusCode: 403,
    });

    await expect(
      processCashDrawerAction({
        action: "paid_out",
        amountInput: "15",
        reason: "Petty cash",
        drawer,
        context: {
          branchId: 1,
          terminalId: 2,
          employeeId: 3,
          resolveSessionId: async () => 10,
        },
        syncMovement: vi.fn(async () => {
          throw rejection;
        }),
        enqueueMovement,
      }),
    ).rejects.toBe(rejection);

    expect(drawer.paidOut).not.toHaveBeenCalled();
    expect(enqueueMovement).not.toHaveBeenCalled();
  });

  it("does not sync opening float as a cash movement", async () => {
    const drawer = actions();
    const syncMovement = vi.fn();

    await processCashDrawerAction({
      action: "float",
      amountInput: "100",
      reason: "",
      drawer,
      context,
      syncMovement,
      enqueueMovement: vi.fn(),
    });

    expect(syncMovement).not.toHaveBeenCalled();
  });
});
