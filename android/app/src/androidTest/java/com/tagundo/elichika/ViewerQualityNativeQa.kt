package com.tagundo.elichika

import android.webkit.WebView
import androidx.test.platform.app.InstrumentationRegistry
import org.json.JSONObject
import org.json.JSONTokener
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference

/** Shared assertions for the real product renderer, never a substitute renderer. */
object ViewerQualityNativeQa {
    fun exercise(webView: WebView, prefix: String, sourceTriangles: Int, drawTriangles: Int,
                 leaveEnabled: Boolean = true, centerRgb: IntArray? = null): JSONObject {
        val instrumentation = InstrumentationRegistry.getInstrumentation()
        val helper = instrumentation.context.assets.open("viewer_quality_qa.js").bufferedReader().use { it.readText() }
        evaluate(webView, helper + "\nreturn JSON.stringify({ok:!!window.ViewerQualityQa&&!!window.ViewerRigQa});")
        val control = evaluate(webView, "return JSON.stringify(ViewerQualityQa.control());")
        assertTrue("The high-quality label must provide an unobstructed 44 px target: $control",
            !control.getBoolean("disabled") && control.getBoolean("withinViewport") && control.getBoolean("hittable") &&
                control.getDouble("width") >= 44 && control.getDouble("height") >= 44)

        setQuality(webView, false)
        val before = state(webView)
        assertFalse("Use a paused pose for exact quality round trips", before.getJSONObject("rig").getBoolean("playing"))
        val standard = frame(webView, prefix + "-standard")
        val standardDriver = driver(webView)
        assertDimensions(standardDriver, false)
        assertFrame(standard, sourceTriangles, drawTriangles, centerRgb)
        evaluate(webView, "ViewerQualityQa.clear();return JSON.stringify({ok:true});")

        setQuality(webView, true)
        val enabled = state(webView)
        assertStableModel(before, enabled)
        val high = frame(webView, prefix + "-hq")
        val highDriver = driver(webView)
        assertDimensions(highDriver, true)
        assertFrame(high, sourceTriangles, drawTriangles, centerRgb)

        setQuality(webView, false)
        assertStableModel(before, state(webView))
        val standardReturn = frame(webView, compare = prefix + "-standard")
        assertExact("Disabling HQ must restore the same standard framebuffer", standardReturn)
        assertDimensions(driver(webView), false)
        setQuality(webView, true)
        assertExact("Re-enabling HQ must reproduce the same native-resolution framebuffer",
            frame(webView, compare = prefix + "-hq"))

        val prior = state(webView)
        val priorQuality = prior.getJSONObject("quality")
        action(webView, "document.getElementById('viewer-expand').click();")
        val resized = waitFor(webView, "the HQ renderbuffer to follow the actual viewer resize") {
            "var s=ModelViewer.getState(),q=s.quality;return JSON.stringify({ok:q&&q.ready&&q.enabled&&" +
                "(q.width!==${priorQuality.getInt("width")}||q.height!==${priorQuality.getInt("height")}),state:s});"
        }
        assertStableModel(prior, resized.getJSONObject("state"))
        assertDimensions(driver(webView), true)
        action(webView, "document.getElementById('viewer-expand').click();")
        waitFor(webView, "the original HQ renderbuffer dimensions") {
            "var q=ModelViewer.getState().quality;return JSON.stringify({ok:q&&q.ready&&q.width===${priorQuality.getInt("width")}" +
                "&&q.height===${priorQuality.getInt("height")}});"
        }
        assertStableModel(before, state(webView))
        val afterResize = frame(webView, compare = prefix + "-hq")
        assertExact("Returning from resize must restore the exact HQ frame", afterResize)

        action(webView, "ViewerRigQa.loseContext();")
        waitFor(webView, "actual HQ context loss") {
            "var s=ModelViewer.getState();return JSON.stringify({ok:s.contextLost&&" +
                "document.getElementById('viewer-high-quality').checked,state:s});"
        }
        action(webView, "ViewerRigQa.restoreContext();")
        waitFor(webView, "the real HQ context and multisample resources to recover") {
            "var s=ModelViewer.getState(),q=s.quality;return JSON.stringify({ok:s.loaded&&!s.loading&&!s.contextLost&&" +
                "q&&q.enabled&&q.ready,state:s});"
        }
        assertStableModel(before, state(webView))
        val recovered = frame(webView, compare = prefix + "-hq")
        assertExact("Context recovery must rebuild the same HQ image", recovered)
        val recoveredDriver = driver(webView)
        assertDimensions(recoveredDriver, true)
        assertFrame(recovered, sourceTriangles, drawTriangles, centerRgb)
        setQuality(webView, false)
        assertExact("The standard image must remain exact after HQ context restoration",
            frame(webView, compare = prefix + "-standard"))
        val rejected = if (highDriver.getJSONObject("quality").getString("mode") == "multisample") {
            val injected = evaluate(webView, "return JSON.stringify(ViewerQualityQa.rejectAllocations());")
            val fallbackReport = try {
                // Deliberately inspect the first rendered frame. Polling a probe
                // until it succeeds could consume and hide an earlier GL error.
                val immediate = evaluate(webView, "ViewerQualityQa.set(true);" +
                    "var p=ModelViewer.renderProbe(33);return JSON.stringify({probe:p,state:ModelViewer.getState(),driver:ViewerQualityQa.snapshot()});")
                assertTrue(immediate.getJSONObject("probe").getBoolean("ready"))
                assertEquals("Rejected owned storage must not leak a GL error into the first fallback frame", 0,
                    immediate.getJSONObject("probe").getInt("glError"))
                val fallback = immediate.getJSONObject("state").getJSONObject("quality")
                assertEquals("multisample-framebuffer-unavailable", fallback.getString("fallbackReason"))
                assertTrue(fallback.getBoolean("enabled"))
                assertFalse(fallback.getBoolean("targetAllocated"))
                assertTrue("Rejected owned storage must retain its error diagnostics", fallback.getJSONArray("allocationErrors").length() >= 1)
                assertTrue(fallback.getInt("recoveredAllocationErrors") >= 1)
                assertTrue(fallback.getInt("allocationFailures") >= 1)
                assertStableModel(before, immediate.getJSONObject("state"))
                assertDimensions(immediate.getJSONObject("driver"), true)
                assertFrame(frame(webView), sourceTriangles, drawTriangles, centerRgb)
                immediate.put("faultInjection", injected)
            } finally {
                action(webView, "ViewerQualityQa.set(false);ViewerQualityQa.restoreAllocations();")
            }
            setQuality(webView, true)
            assertDimensions(driver(webView), true)
            assertExact("Normal HQ allocation must recover exactly after the rejected attempt",
                frame(webView, compare = prefix + "-hq"))
            setQuality(webView, false)
            assertExact("A rejected allocation must not poison the later standard frame",
                frame(webView, compare = prefix + "-standard"))
            fallbackReport
        } else JSONObject().put("skipped", "No supported multisample render target on this driver")
        if (leaveEnabled) setQuality(webView, true)
        assertStableModel(before, state(webView))

        return JSONObject().put("control", control).put("standard", standard).put("standardDriver", standardDriver)
            .put("high", high).put("highDriver", highDriver).put("standardReturn", standardReturn)
            .put("afterResize", afterResize).put("recovered", recovered).put("recoveredDriver", recoveredDriver)
            .put("rejectedAllocation", rejected)
            .put("retainedModelId", before.getString("modelId")).put("retainedRigAndCamera", true)
            .put("hqReferenceFrame", prefix + "-hq").put("finalState", state(webView))
    }

    fun assertStableModel(before: JSONObject, after: JSONObject) {
        assertEquals("Quality must not reload the model", before.getString("modelId"), after.getString("modelId"))
        assertEquals("Quality must preserve the selected inputs", before.getJSONArray("loadedPaths").toString(), after.getJSONArray("loadedPaths").toString())
        assertEquals("Quality must preserve the same camera", before.getJSONObject("camera").toString(), after.getJSONObject("camera").toString())
        val first = before.getJSONObject("rig"); val second = after.getJSONObject("rig")
        for (key in listOf("selectedClip", "defaultClip", "playing", "time", "body", "nodeAdjustments", "savedNodes", "expressions", "bones")) {
            assertEquals("Quality must preserve the actual rig's $key", first.get(key).toString(), second.get(key).toString())
        }
    }

    private fun assertDimensions(observation: JSONObject, high: Boolean) {
        val q = observation.getJSONObject("quality")
        assertTrue("The actual quality pipeline must be ready; observation=$observation", q.getBoolean("ready"))
        assertEquals(high, q.getBoolean("enabled"))
        assertEquals(observation.getInt("drawingBufferWidth"), q.getInt("width"))
        assertEquals(observation.getInt("drawingBufferHeight"), q.getInt("height"))
        assertEquals(observation.getInt("canvasWidth"), q.getInt("width"))
        assertEquals(observation.getInt("canvasHeight"), q.getInt("height"))
        val nativeDpr = observation.getDouble("nativePixelRatio")
        assertEquals(nativeDpr, q.getDouble("nativePixelRatio"), 1e-6)
        val expectedDpr = if (high) nativeDpr else minOf(nativeDpr, 1.75)
        if (!q.getBoolean("dimensionLimited")) assertEquals(expectedDpr, q.getDouble("effectivePixelRatio"), 1e-6)
        assertEquals(kotlin.math.floor(q.getDouble("cssWidth") * q.getDouble("effectivePixelRatio")).toInt(), q.getInt("width"))
        assertEquals(kotlin.math.floor(q.getDouble("cssHeight") * q.getDouble("effectivePixelRatio")).toInt(), q.getInt("height"))
        if (high) {
            assertEquals("The native SIFAS option supplies the two-sample baseline", 2, q.getInt("gameSamples"))
            val samples = q.getInt("actualSamples")
            assertEquals(samples, q.getInt("effectiveSamples"))
            if (q.getString("mode") == "multisample") {
                assertTrue(q.getBoolean("targetAllocated"))
                assertTrue("HQ must preserve any stronger existing context AA", samples >= maxOf(2, q.getInt("contextSamples")))
                val allocations = observation.getJSONArray("allocations")
                assertTrue("A real driver allocation must agree with the reported samples/dimensions",
                    (0 until allocations.length()).any { index ->
                        val entry = allocations.getJSONObject(index)
                        entry.getInt("requestedSamples") == q.getInt("targetRequestedSamples") &&
                            entry.getInt("actualSamples") == samples && entry.getInt("width") == q.getInt("width") && entry.getInt("height") == q.getInt("height")
                    })
            } else {
                assertTrue("A supported fallback must explain why it uses context AA", q.getString("fallbackReason").isNotBlank())
                assertEquals(observation.getInt("contextSamples"), samples)
            }
        } else {
            assertEquals(0, q.getInt("gameSamples"))
            assertFalse("Standard mode must release its extra render target", q.getBoolean("targetAllocated"))
        }
    }

    private fun assertFrame(frame: JSONObject, source: Int, draw: Int, centerRgb: IntArray?) {
        assertEquals(0, frame.getInt("glError"))
        val probe = frame.getJSONObject("probe")
        assertEquals(source, probe.getInt("sourceTriangles"))
        assertEquals("Composition geometry must not inflate model triangle counts", draw, probe.getInt("triangles"))
        assertTrue("The real model must remain visible", frame.getInt("foregroundPixels") >= 64)
        if (centerRgb != null) for (i in centerRgb.indices) {
            assertTrue("The source gamma color must remain unchanged by HQ composition", kotlin.math.abs(frame.getJSONArray("centerRGB").getInt(i) - centerRgb[i]) <= 3)
        }
    }

    private fun assertExact(message: String, frame: JSONObject) {
        val diff = frame.getJSONObject("difference")
        assertEquals(message, 0, diff.getInt("changedPixels"))
        assertEquals(message, 0, diff.getInt("maximumChannelDifference"))
        assertEquals(message, 0.0, diff.getDouble("cameraMaximumDifference"), 1e-7)
        assertEquals(0, frame.getInt("glError"))
    }

    private fun setQuality(webView: WebView, enabled: Boolean) {
        action(webView, "ViewerQualityQa.set($enabled);")
        waitFor(webView, "the requested high-quality setting") {
            "var s=ModelViewer.getState(),q=s.quality,p=ModelViewer.renderProbe(33);" +
                "return JSON.stringify({ok:q&&q.ready&&q.enabled===$enabled&&p.ready&&p.glError===0,state:s,probe:p});"
        }
    }

    private fun driver(webView: WebView) = evaluate(webView, "return JSON.stringify(ViewerQualityQa.snapshot());")
    private fun state(webView: WebView) = evaluate(webView, "return JSON.stringify(ModelViewer.getState());")
    private fun frame(webView: WebView, save: String? = null, compare: String? = null) = evaluate(webView,
        "return JSON.stringify(ViewerRigQa.capture(${save?.let { JSONObject.quote(it) } ?: "null"},${compare?.let { JSONObject.quote(it) } ?: "null"}));")
    private fun action(webView: WebView, code: String) { evaluate(webView, code + "\nreturn JSON.stringify({ok:true});") }
    private fun waitFor(webView: WebView, description: String, script: () -> String): JSONObject {
        val end = System.nanoTime() + TimeUnit.SECONDS.toNanos(90)
        var last = JSONObject()
        while (System.nanoTime() < end) {
            last = evaluate(webView, script())
            if (last.optBoolean("ok")) return last
            Thread.sleep(100)
        }
        throw AssertionError("Timed out waiting for $description: $last")
    }
    private fun evaluate(webView: WebView, body: String): JSONObject {
        val latch = CountDownLatch(1); val encoded = AtomicReference<String>()
        InstrumentationRegistry.getInstrumentation().runOnMainSync {
            webView.evaluateJavascript("(function(){try{$body}catch(e){return JSON.stringify({qualityQaError:String(e.stack||e)});}})();") {
                encoded.set(it); latch.countDown()
            }
        }
        assertTrue("The actual WebView must answer quality QA", latch.await(30, TimeUnit.SECONDS))
        val value = JSONTokener(encoded.get()).nextValue()
        val result = if (value is String) JSONObject(value) else value as JSONObject
        assertFalse("Quality QA must use functioning product APIs: $result", result.has("qualityQaError"))
        return result
    }
}
