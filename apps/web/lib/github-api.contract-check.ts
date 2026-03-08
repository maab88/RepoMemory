import type { GitHubCallbackSuccessResponse, StartGitHubConnectResponse } from "@repomemory/contracts";
import { completeGitHubCallback, startGitHubConnect } from "@/lib/github-api";

type IsEqual<A, B> =
  (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;

type _StartConnectType = Assert<IsEqual<Awaited<ReturnType<typeof startGitHubConnect>>, StartGitHubConnectResponse["data"]>>;
type _CallbackType = Assert<IsEqual<Awaited<ReturnType<typeof completeGitHubCallback>>, GitHubCallbackSuccessResponse["data"]>>;