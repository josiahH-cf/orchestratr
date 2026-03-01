<!-- generated-from-metaprompt -->
You are helping structure a raw idea into actionable GitHub Issues.

The idea is: $ARGUMENTS

Do the following:

1. Determine whether this is a single feature or a batch of related features.
   - If it is a single feature, produce ONE issue.
   - If it is a batch or epic, decompose it into independently actionable issues. Each issue must stand alone — no implicit ordering. If issues depend on each other, state the dependency explicitly in the dependent issue's Notes section.

2. For each issue, produce the following and nothing else:

   ---
   **Title:** [concise imperative title]

   **Idea:** [1–3 sentences: what and why]

   **Desired Outcome:** [What should be true when this is done]

   **Sizing Guess:** [S / M / L / XL — if XL, note that it should be split further during scoping]

   **Notes:** [Context, links, prior art, dependencies on other issues if any]

   **Labels:** type:feature (or type:bug), status:idea, size:[S/M/L/XL]
   ---

3. If you produced multiple issues, list them in a suggested scoping order (dependency-first).

Do not write code. Do not create specs. Do not scope the feature. The sole output of this phase is one or more issue bodies ready to be filed.
