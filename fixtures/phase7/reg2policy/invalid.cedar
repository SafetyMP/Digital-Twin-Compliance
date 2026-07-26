// Intentionally invalid Cedar for CI reject path
permit (
  principal in Role::"Analyst"
  action == Action::"viewAlert"
  resource
;
