using System;
using System.Threading.Tasks;
using Newtonsoft.Json.Linq;
using UnityEditor;
using UnityEditor.Compilation;

namespace UnityCliConnector.Tools
{
    [UnityCliTool(Description = "Refresh Unity assets and optionally request script compilation.")]
    public static class RefreshUnity
    {
        private const int DefaultWaitTimeoutSeconds = 60;

        public class Parameters
        {
            [ToolParameter("Refresh mode: if_dirty (default) or force")]
            public string Mode { get; set; }

            [ToolParameter("Scope: all (default) or specific path")]
            public string Scope { get; set; }

            [ToolParameter("Compile mode: none (default), request, or wait")]
            public string Compile { get; set; }

            [ToolParameter("Wait until Unity is fully ready after refresh")]
            public bool WaitForReady { get; set; }
        }

        public static async Task<object> HandleCommand(JObject @params)
        {
            string mode = @params?["mode"]?.ToString() ?? "if_dirty";
            string scope = @params?["scope"]?.ToString() ?? "all";
            string compile = @params?["compile"]?.ToString() ?? "none";
            bool waitForReady = ParamCoercion.CoerceBool(@params?["wait_for_ready"], false);

            bool refreshTriggered = false;
            bool compileRequested = false;

            bool shouldRefresh = string.Equals(mode, "force", StringComparison.OrdinalIgnoreCase)
                                 || string.Equals(mode, "if_dirty", StringComparison.OrdinalIgnoreCase);

            if (shouldRefresh && !string.Equals(scope, "scripts", StringComparison.OrdinalIgnoreCase))
            {
                AssetDatabase.Refresh(ImportAssetOptions.ForceUpdate | ImportAssetOptions.ForceSynchronousImport);
                refreshTriggered = true;
            }

            if (string.Equals(compile, "request", StringComparison.OrdinalIgnoreCase)
                || string.Equals(compile, "wait", StringComparison.OrdinalIgnoreCase))
            {
                CompilationPipeline.RequestScriptCompilation();
                compileRequested = true;
            }

            if (string.Equals(scope, "all", StringComparison.OrdinalIgnoreCase) && !refreshTriggered)
            {
                AssetDatabase.Refresh(ImportAssetOptions.ForceSynchronousImport);
                refreshTriggered = true;
            }

            bool shouldWait = compileRequested
                ? string.Equals(compile, "wait", StringComparison.OrdinalIgnoreCase)
                : waitForReady;

            if (shouldWait)
            {
                await WaitForCompileAndReady(TimeSpan.FromSeconds(DefaultWaitTimeoutSeconds));
            }

            string resultingState = EditorApplication.isCompiling
                ? "compiling"
                : (EditorApplication.isUpdating ? "asset_import" : "idle");

            return new SuccessResponse(shouldWait ? "Refresh completed." : "Refresh requested.", new
            {
                refresh_triggered = refreshTriggered,
                compile_requested = compileRequested,
                waited = shouldWait,
                resulting_state = resultingState,
            });
        }

        private static Task WaitForCompileAndReady(TimeSpan timeout)
        {
            var tcs = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
            var start = DateTime.UtcNow;
            bool sawCompiling = EditorApplication.isCompiling;

            void Tick()
            {
                if (tcs.Task.IsCompleted) { EditorApplication.update -= Tick; return; }
                if ((DateTime.UtcNow - start) > timeout)
                {
                    EditorApplication.update -= Tick;
                    tcs.TrySetException(new TimeoutException());
                    return;
                }

                if (!sawCompiling && EditorApplication.isCompiling)
                    sawCompiling = true;

                if (sawCompiling && !EditorApplication.isCompiling && !EditorApplication.isUpdating)
                {
                    EditorApplication.update -= Tick;
                    tcs.TrySetResult(true);
                    return;
                }

                if (!sawCompiling && !EditorApplication.isCompiling && !EditorApplication.isUpdating)
                {
                    EditorApplication.update -= Tick;
                    tcs.TrySetResult(true);
                }
            }

            EditorApplication.update += Tick;
            try { EditorApplication.QueuePlayerLoopUpdate(); } catch { }
            return tcs.Task;
        }
    }
}
