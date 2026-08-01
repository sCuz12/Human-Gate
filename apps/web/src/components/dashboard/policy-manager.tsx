"use client";

import { useEffect, useMemo, useState, type FormEvent } from "react";
import type { Session } from "@supabase/supabase-js";

import {
  createPolicy,
  deletePolicy,
  listPolicies,
  updatePolicy,
  type Policy,
  type PolicyCondition,
} from "@/lib/api/policies";
import type { Workspace } from "@/lib/api/workspaces";

type PolicyManagerProps = {
  session: Session;
  workspace: Workspace;
};

const effectLabels: Record<Policy["effect"], string> = {
  allow: "Allow automatically",
  require_approval: "Require approval",
  block: "Block",
};

const fieldLabels: Record<PolicyCondition["field"], string> = {
  "action.type": "Action type",
  "source.platform": "Source platform",
  "context.reversible": "Reversible",
};

export function PolicyManager({ session, workspace }: PolicyManagerProps) {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [name, setName] = useState("Refund approval");
  const [actionType, setActionType] = useState("customer.refund");
  const [sourcePlatform, setSourcePlatform] = useState("n8n");
  const [effect, setEffect] = useState<Policy["effect"]>("require_approval");
  const [deadlineMinutes, setDeadlineMinutes] = useState(5);
  const [priority, setPriority] = useState(100);
  const [isActive, setIsActive] = useState(true);
  const [editingPolicyID, setEditingPolicyID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deletingPolicyID, setDeletingPolicyID] = useState<string | null>(null);

  useEffect(() => {
    void loadPolicies();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspace.id]);

  async function loadPolicies() {
    try {
      setIsLoading(true);
      const response = await listPolicies(session, workspace.id);
      setPolicies(response.policies);
      setError(null);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Policies could not be loaded.");
    } finally {
      setIsLoading(false);
    }
  }

  const previewConditions = useMemo(() => {
    const conditions: PolicyCondition[] = [];
    if (actionType.trim()) {
      conditions.push({
        field: "action.type",
        operator: "equals",
        value: actionType.trim(),
      });
    }
    if (sourcePlatform.trim()) {
      conditions.push({
        field: "source.platform",
        operator: "equals",
        value: sourcePlatform.trim(),
      });
    }
    return conditions;
  }, [actionType, sourcePlatform]);

  async function submitPolicy(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);

    try {
      const input = {
        workspaceID: workspace.id,
        name,
        priority,
        isActive,
        conditions: previewConditions,
        effect,
        deadlineSeconds: effect === "require_approval" ? deadlineMinutes * 60 : 0,
      };

      const response = editingPolicyID
        ? await updatePolicy(session, {
            ...input,
            policyID: editingPolicyID,
          })
        : await createPolicy(session, input);

      setPolicies((current) => {
        const next = editingPolicyID
          ? current.map((policy) => (policy.id === response.policy.id ? response.policy : policy))
          : [...current, response.policy];
        return next.sort((a, b) => a.priority - b.priority);
      });
      resetForm();
      setError(null);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Policy could not be saved.");
    } finally {
      setIsSubmitting(false);
    }
  }

  function editPolicy(policy: Policy) {
    setEditingPolicyID(policy.id);
    setName(policy.name);
    setActionType(stringConditionValue(policy, "action.type") ?? "");
    setSourcePlatform(stringConditionValue(policy, "source.platform") ?? "");
    setEffect(policy.effect);
    setDeadlineMinutes(policy.deadline_seconds ? Math.max(1, Math.round(policy.deadline_seconds / 60)) : 5);
    setPriority(policy.priority);
    setIsActive(policy.is_active);
    setError(null);
  }

  function resetForm() {
    setEditingPolicyID(null);
    setName("Refund approval");
    setActionType("customer.refund");
    setSourcePlatform("n8n");
    setEffect("require_approval");
    setDeadlineMinutes(5);
    setPriority(100);
    setIsActive(true);
  }

  async function removePolicy(policy: Policy) {
    const confirmed = window.confirm(`Delete "${policy.name}"? Existing approval history will remain.`);
    if (!confirmed) {
      return;
    }

    setDeletingPolicyID(policy.id);
    try {
      await deletePolicy(session, {
        workspaceID: workspace.id,
        policyID: policy.id,
      });
      setPolicies((current) => current.filter((item) => item.id !== policy.id));
      if (editingPolicyID === policy.id) {
        resetForm();
      }
      setError(null);
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : "Policy could not be deleted.");
    } finally {
      setDeletingPolicyID(null);
    }
  }

  return (
    <section className="mt-5 rounded-lg border border-black/10 bg-white p-4">
      <div className="flex flex-col gap-1">
        <p className="text-sm font-semibold text-[#15110d]">
          {editingPolicyID ? "Edit policy" : "Create policy"}
        </p>
        <p className="text-sm text-black/55">
          Route workflow actions, decide the outcome, and set approval deadlines.
        </p>
      </div>

      {error ? (
        <div className="mt-4 rounded-md border border-[#b74b2a]/20 bg-[#fff2ec] p-3 text-sm text-[#8d3419]">
          {error}
        </div>
      ) : null}

      <form className="mt-4 grid gap-3" onSubmit={(event) => void submitPolicy(event)}>
        <div className="grid gap-3 md:grid-cols-2">
          <label className="text-sm">
            <span className="font-medium text-black/70">Policy name</span>
            <input
              className="mt-1 w-full rounded-md border border-black/10 px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
              onChange={(event) => setName(event.target.value)}
              required
              value={name}
            />
          </label>
          <label className="text-sm">
            <span className="font-medium text-black/70">Priority</span>
            <input
              className="mt-1 w-full rounded-md border border-black/10 px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
              min={1}
              onChange={(event) => setPriority(Number(event.target.value))}
              type="number"
              value={priority}
            />
          </label>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <label className="text-sm">
            <span className="font-medium text-black/70">Action type</span>
            <input
              className="mt-1 w-full rounded-md border border-black/10 px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
              onChange={(event) => setActionType(event.target.value)}
              placeholder="customer.refund"
              value={actionType}
            />
          </label>
          <label className="text-sm">
            <span className="font-medium text-black/70">Source platform</span>
            <input
              className="mt-1 w-full rounded-md border border-black/10 px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
              onChange={(event) => setSourcePlatform(event.target.value)}
              placeholder="n8n"
              value={sourcePlatform}
            />
          </label>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <label className="text-sm">
            <span className="font-medium text-black/70">Effect</span>
            <select
              className="mt-1 w-full rounded-md border border-black/10 bg-white px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15"
              onChange={(event) => setEffect(event.target.value as Policy["effect"])}
              value={effect}
            >
              <option value="require_approval">Require approval</option>
              <option value="allow">Allow automatically</option>
              <option value="block">Block</option>
            </select>
          </label>
          <label className="text-sm">
            <span className="font-medium text-black/70">Deadline minutes</span>
            <input
              className="mt-1 w-full rounded-md border border-black/10 px-3 py-2 outline-none focus:border-[#1f6f78] focus:ring-2 focus:ring-[#1f6f78]/15 disabled:bg-black/5"
              disabled={effect !== "require_approval"}
              min={1}
              onChange={(event) => setDeadlineMinutes(Number(event.target.value))}
              type="number"
              value={deadlineMinutes}
            />
          </label>
        </div>

        <label className="inline-flex items-center gap-2 text-sm text-black/70">
          <input
            checked={isActive}
            className="h-4 w-4"
            onChange={(event) => setIsActive(event.target.checked)}
            type="checkbox"
          />
          Active immediately
        </label>

        <button
          className="rounded-md bg-[#15110d] px-4 py-3 text-sm font-semibold text-white transition hover:bg-black disabled:cursor-not-allowed disabled:opacity-50"
          disabled={isSubmitting || previewConditions.length === 0}
          type="submit"
        >
          {isSubmitting ? "Saving policy..." : editingPolicyID ? "Save changes" : "Create policy"}
        </button>
        {editingPolicyID ? (
          <button
            className="rounded-md border border-black/10 px-4 py-3 text-sm font-semibold text-black/65 transition hover:border-[#1f6f78] hover:text-[#1f6f78]"
            onClick={resetForm}
            type="button"
          >
            Cancel editing
          </button>
        ) : null}
      </form>

      <div className="mt-5 border-t border-black/10 pt-4">
        {isLoading ? (
          <p className="text-sm text-black/55">Loading policies...</p>
        ) : policies.length === 0 ? (
          <p className="text-sm text-black/55">No policies created yet.</p>
        ) : (
          <div className="grid gap-3">
            {policies.map((policy) => (
              <article key={policy.id} className="rounded-md border border-black/10 bg-[#fbf8f2] p-3">
                <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                  <div>
                    <p className="font-medium text-[#15110d]">{policy.name}</p>
                    <p className="mt-1 text-sm text-black/55">
                      {effectLabels[policy.effect]} · priority {policy.priority}
                    </p>
                  </div>
                  <span className="w-fit rounded-md border border-black/10 bg-white px-2 py-1 text-xs font-medium text-black/65">
                    {policy.is_active ? "Active" : "Inactive"}
                  </span>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {policy.conditions.map((condition, index) => (
                    <span
                      className="rounded-md bg-white px-2 py-1 text-xs text-black/65"
                      key={`${policy.id}-${condition.field}-${index}`}
                    >
                      {fieldLabels[condition.field]} = {String(condition.value)}
                    </span>
                  ))}
                  {policy.deadline_seconds ? (
                    <span className="rounded-md bg-[#eef8f9] px-2 py-1 text-xs text-[#13505a]">
                      Deadline {Math.round(policy.deadline_seconds / 60)} min
                    </span>
                  ) : null}
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <button
                    className="rounded-md border border-black/10 bg-white px-3 py-2 text-xs font-semibold text-black/65 transition hover:border-[#1f6f78] hover:text-[#1f6f78]"
                    onClick={() => editPolicy(policy)}
                    type="button"
                  >
                    Edit
                  </button>
                  <button
                    className="rounded-md border border-[#b74b2a]/30 bg-white px-3 py-2 text-xs font-semibold text-[#8d3419] transition hover:bg-[#fff2ec] disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={deletingPolicyID === policy.id}
                    onClick={() => void removePolicy(policy)}
                    type="button"
                  >
                    {deletingPolicyID === policy.id ? "Deleting..." : "Delete"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

function stringConditionValue(policy: Policy, field: PolicyCondition["field"]) {
  const condition = policy.conditions.find((item) => item.field === field);
  return typeof condition?.value === "string" ? condition.value : null;
}
