(function(){const c=document.createElement("link").relList;if(c&&c.supports&&c.supports("modulepreload"))return;for(const h of document.querySelectorAll('link[rel="modulepreload"]'))r(h);new MutationObserver(h=>{for(const m of h)if(m.type==="childList")for(const g of m.addedNodes)g.tagName==="LINK"&&g.rel==="modulepreload"&&r(g)}).observe(document,{childList:!0,subtree:!0});function u(h){const m={};return h.integrity&&(m.integrity=h.integrity),h.referrerPolicy&&(m.referrerPolicy=h.referrerPolicy),h.crossOrigin==="use-credentials"?m.credentials="include":h.crossOrigin==="anonymous"?m.credentials="omit":m.credentials="same-origin",m}function r(h){if(h.ep)return;h.ep=!0;const m=u(h);fetch(h.href,m)}})();function K1(o){return o&&o.__esModule&&Object.prototype.hasOwnProperty.call(o,"default")?o.default:o}var pd={exports:{}},Vi={};var ym;function F1(){if(ym)return Vi;ym=1;var o=Symbol.for("react.transitional.element"),c=Symbol.for("react.fragment");function u(r,h,m){var g=null;if(m!==void 0&&(g=""+m),h.key!==void 0&&(g=""+h.key),"key"in h){m={};for(var v in h)v!=="key"&&(m[v]=h[v])}else m=h;return h=m.ref,{$$typeof:o,type:r,key:g,ref:h!==void 0?h:null,props:m}}return Vi.Fragment=c,Vi.jsx=u,Vi.jsxs=u,Vi}var xm;function J1(){return xm||(xm=1,pd.exports=F1()),pd.exports}var a=J1();const Cd=""+new URL("service-logo-BZK6IUBf.webp",import.meta.url).href;var fd={exports:{}},ge={};var wm;function W1(){if(wm)return ge;wm=1;var o=Symbol.for("react.transitional.element"),c=Symbol.for("react.portal"),u=Symbol.for("react.fragment"),r=Symbol.for("react.strict_mode"),h=Symbol.for("react.profiler"),m=Symbol.for("react.consumer"),g=Symbol.for("react.context"),v=Symbol.for("react.forward_ref"),b=Symbol.for("react.suspense"),C=Symbol.for("react.memo"),x=Symbol.for("react.lazy"),y=Symbol.for("react.activity"),w=Symbol.for("react.view_transition"),A=Symbol.iterator;function E(j){return j===null||typeof j!="object"?null:(j=A&&j[A]||j["@@iterator"],typeof j=="function"?j:null)}var _={isMounted:function(){return!1},enqueueForceUpdate:function(){},enqueueReplaceState:function(){},enqueueSetState:function(){}},N=Object.assign,D={};function H(j,I,ee){this.props=j,this.context=I,this.refs=D,this.updater=ee||_}H.prototype.isReactComponent={},H.prototype.setState=function(j,I){if(typeof j!="object"&&typeof j!="function"&&j!=null)throw Error("takes an object of state variables to update or a function which returns an object of state variables.");this.updater.enqueueSetState(this,j,I,"setState")},H.prototype.forceUpdate=function(j){this.updater.enqueueForceUpdate(this,j,"forceUpdate")};function G(){}G.prototype=H.prototype;function Q(j,I,ee){this.props=j,this.context=I,this.refs=D,this.updater=ee||_}var Y=Q.prototype=new G;Y.constructor=Q,N(Y,H.prototype),Y.isPureReactComponent=!0;var S=Array.isArray;function U(){}var Z={H:null,A:null,T:null,S:null},$=Object.prototype.hasOwnProperty;function oe(j,I,ee){var ie=ee.ref;return{$$typeof:o,type:j,key:I,ref:ie!==void 0?ie:null,props:ee}}function be(j,I){return oe(j.type,I,j.props)}function re(j){return typeof j=="object"&&j!==null&&j.$$typeof===o}function J(j){var I={"=":"=0",":":"=2"};return"$"+j.replace(/[=:]/g,function(ee){return I[ee]})}var ve=/\/+/g;function W(j,I){return typeof j=="object"&&j!==null&&j.key!=null?J(""+j.key):I.toString(36)}function F(j){switch(j.status){case"fulfilled":return j.value;case"rejected":throw j.reason;default:switch(typeof j.status=="string"?j.then(U,U):(j.status="pending",j.then(function(I){j.status==="pending"&&(j.status="fulfilled",j.value=I)},function(I){j.status==="pending"&&(j.status="rejected",j.reason=I)})),j.status){case"fulfilled":return j.value;case"rejected":throw j.reason}}throw j}function ne(j,I,ee,ie,fe){var ye=typeof j;(ye==="undefined"||ye==="boolean")&&(j=null);var xe=!1;if(j===null)xe=!0;else switch(ye){case"bigint":case"string":case"number":xe=!0;break;case"object":switch(j.$$typeof){case o:case c:xe=!0;break;case x:return xe=j._init,ne(xe(j._payload),I,ee,ie,fe)}}if(xe)return fe=fe(j),xe=ie===""?"."+W(j,0):ie,S(fe)?(ee="",xe!=null&&(ee=xe.replace(ve,"$&/")+"/"),ne(fe,I,ee,"",function(ut){return ut})):fe!=null&&(re(fe)&&(fe=be(fe,ee+(fe.key==null||j&&j.key===fe.key?"":(""+fe.key).replace(ve,"$&/")+"/")+xe)),I.push(fe)),1;xe=0;var ae=ie===""?".":ie+":";if(S(j))for(var he=0;he<j.length;he++)ie=j[he],ye=ae+W(ie,he),xe+=ne(ie,I,ee,ye,fe);else if(he=E(j),typeof he=="function")for(j=he.call(j),he=0;!(ie=j.next()).done;)ie=ie.value,ye=ae+W(ie,he++),xe+=ne(ie,I,ee,ye,fe);else if(ye==="object"){if(typeof j.then=="function")return ne(F(j),I,ee,ie,fe);throw I=String(j),Error("Objects are not valid as a React child (found: "+(I==="[object Object]"?"object with keys {"+Object.keys(j).join(", ")+"}":I)+"). If you meant to render a collection of children, use an array instead.")}return xe}function le(j,I,ee){if(j==null)return j;var ie=[],fe=0;return ne(j,ie,"","",function(ye){return I.call(ee,ye,fe++)}),ie}function je(j){if(j._status===-1){var I=j._result,ee=I();ee.then(function(ie){(j._status===0||j._status===-1)&&(j._status=1,j._result=ie,ee.status===void 0&&(ee.status="fulfilled",ee.value=ie))},function(ie){(j._status===0||j._status===-1)&&(j._status=2,j._result=ie,ee.status===void 0&&(ee.status="rejected",ee.reason=ie))}),j._status===-1&&(j._status=0,j._result=ee)}if(j._status===1)return j._result.default;throw j._result}var Ae=typeof reportError=="function"?reportError:function(j){if(typeof window=="object"&&typeof window.ErrorEvent=="function"){var I=new window.ErrorEvent("error",{bubbles:!0,cancelable:!0,message:typeof j=="object"&&j!==null&&typeof j.message=="string"?String(j.message):String(j),error:j});if(!window.dispatchEvent(I))return}else if(typeof process=="object"&&typeof process.emit=="function"){process.emit("uncaughtException",j);return}console.error(j)};function Qe(j){var I=Z.T,ee={};ee.types=I!==null?I.types:null,Z.T=ee;try{var ie=j(),fe=Z.S;fe!==null&&fe(ee,ie),typeof ie=="object"&&ie!==null&&typeof ie.then=="function"&&ie.then(U,Ae)}catch(ye){Ae(ye)}finally{I!==null&&ee.types!==null&&(I.types=ee.types),Z.T=I}}function Ze(j){var I=Z.T;if(I!==null){var ee=I.types;ee===null?I.types=[j]:ee.indexOf(j)===-1&&ee.push(j)}else Qe(Ze.bind(null,j))}var dt={map:le,forEach:function(j,I,ee){le(j,function(){I.apply(this,arguments)},ee)},count:function(j){var I=0;return le(j,function(){I++}),I},toArray:function(j){return le(j,function(I){return I})||[]},only:function(j){if(!re(j))throw Error("React.Children.only expected to receive a single React element child.");return j}};return ge.Activity=y,ge.Children=dt,ge.Component=H,ge.Fragment=u,ge.Profiler=h,ge.PureComponent=Q,ge.StrictMode=r,ge.Suspense=b,ge.ViewTransition=w,ge.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE=Z,ge.__COMPILER_RUNTIME={__proto__:null,c:function(j){return Z.H.useMemoCache(j)}},ge.addTransitionType=Ze,ge.cache=function(j){return function(){return j.apply(null,arguments)}},ge.cacheSignal=function(){return null},ge.cloneElement=function(j,I,ee){if(j==null)throw Error("The argument must be a React element, but you passed "+j+".");var ie=N({},j.props),fe=j.key;if(I!=null)for(ye in I.key!==void 0&&(fe=""+I.key),I)!$.call(I,ye)||ye==="key"||ye==="__self"||ye==="__source"||ye==="ref"&&I.ref===void 0||(ie[ye]=I[ye]);var ye=arguments.length-2;if(ye===1)ie.children=ee;else if(1<ye){for(var xe=Array(ye),ae=0;ae<ye;ae++)xe[ae]=arguments[ae+2];ie.children=xe}return oe(j.type,fe,ie)},ge.createContext=function(j){return j={$$typeof:g,_currentValue:j,_currentValue2:j,_threadCount:0,Provider:null,Consumer:null},j.Provider=j,j.Consumer={$$typeof:m,_context:j},j},ge.createElement=function(j,I,ee){var ie,fe={},ye=null;if(I!=null)for(ie in I.key!==void 0&&(ye=""+I.key),I)$.call(I,ie)&&ie!=="key"&&ie!=="__self"&&ie!=="__source"&&(fe[ie]=I[ie]);var xe=arguments.length-2;if(xe===1)fe.children=ee;else if(1<xe){for(var ae=Array(xe),he=0;he<xe;he++)ae[he]=arguments[he+2];fe.children=ae}if(j&&j.defaultProps)for(ie in xe=j.defaultProps,xe)fe[ie]===void 0&&(fe[ie]=xe[ie]);return oe(j,ye,fe)},ge.createRef=function(){return{current:null}},ge.forwardRef=function(j){return{$$typeof:v,render:j}},ge.isValidElement=re,ge.lazy=function(j){return{$$typeof:x,_payload:{_status:-1,_result:j},_init:je}},ge.memo=function(j,I){return{$$typeof:C,type:j,compare:I===void 0?null:I}},ge.startTransition=Qe,ge.unstable_useCacheRefresh=function(){return Z.H.useCacheRefresh()},ge.use=function(j){return Z.H.use(j)},ge.useActionState=function(j,I,ee){return Z.H.useActionState(j,I,ee)},ge.useCallback=function(j,I){return Z.H.useCallback(j,I)},ge.useContext=function(j){return Z.H.useContext(j)},ge.useDebugValue=function(){},ge.useDeferredValue=function(j,I){return Z.H.useDeferredValue(j,I)},ge.useEffect=function(j,I){return Z.H.useEffect(j,I)},ge.useEffectEvent=function(j){return Z.H.useEffectEvent(j)},ge.useId=function(){return Z.H.useId()},ge.useImperativeHandle=function(j,I,ee){return Z.H.useImperativeHandle(j,I,ee)},ge.useInsertionEffect=function(j,I){return Z.H.useInsertionEffect(j,I)},ge.useLayoutEffect=function(j,I){return Z.H.useLayoutEffect(j,I)},ge.useMemo=function(j,I){return Z.H.useMemo(j,I)},ge.useOptimistic=function(j,I){return Z.H.useOptimistic(j,I)},ge.useReducer=function(j,I,ee){return Z.H.useReducer(j,I,ee)},ge.useRef=function(j){return Z.H.useRef(j)},ge.useState=function(j){return Z.H.useState(j)},ge.useSyncExternalStore=function(j,I,ee){return Z.H.useSyncExternalStore(j,I,ee)},ge.useTransition=function(){return Z.H.useTransition()},ge.version="19.3.0",ge}var bm;function Dd(){return bm||(bm=1,fd.exports=W1()),fd.exports}var f=Dd();const $i=K1(f);var md={exports:{}},Qi={},gd={exports:{}},vd={};var jm;function P1(){return jm||(jm=1,(function(o){function c(F,ne){var le=F.length;F.push(ne);e:for(;0<le;){var je=le-1>>>1,Ae=F[je];if(0<h(Ae,ne))F[je]=ne,F[le]=Ae,le=je;else break e}}function u(F){return F.length===0?null:F[0]}function r(F){if(F.length===0)return null;var ne=F[0],le=F.pop();if(le!==ne){F[0]=le;e:for(var je=0,Ae=F.length,Qe=Ae>>>1;je<Qe;){var Ze=2*(je+1)-1,dt=F[Ze],j=Ze+1,I=F[j];if(0>h(dt,le))j<Ae&&0>h(I,dt)?(F[je]=I,F[j]=le,je=j):(F[je]=dt,F[Ze]=le,je=Ze);else if(j<Ae&&0>h(I,le))F[je]=I,F[j]=le,je=j;else break e}}return ne}function h(F,ne){var le=F.sortIndex-ne.sortIndex;return le!==0?le:F.id-ne.id}if(o.unstable_now=void 0,typeof performance=="object"&&typeof performance.now=="function"){var m=performance;o.unstable_now=function(){return m.now()}}else{var g=Date,v=g.now();o.unstable_now=function(){return g.now()-v}}var b=[],C=[],x=1,y=null,w=3,A=!1,E=!1,_=!1,N=!1,D=typeof setTimeout=="function"?setTimeout:null,H=typeof clearTimeout=="function"?clearTimeout:null,G=typeof setImmediate<"u"?setImmediate:null;function Q(F){for(var ne=u(C);ne!==null;){if(ne.callback===null)r(C);else if(ne.startTime<=F)r(C),ne.sortIndex=ne.expirationTime,c(b,ne);else break;ne=u(C)}}function Y(F){if(_=!1,Q(F),!E)if(u(b)!==null)E=!0,S||(S=!0,re());else{var ne=u(C);ne!==null&&W(Y,ne.startTime-F)}}var S=!1,U=-1,Z=5,$=-1;function oe(){return N?!0:!(o.unstable_now()-$<Z)}function be(){if(N=!1,S){var F=o.unstable_now();$=F;var ne=!0;try{e:{E=!1,_&&(_=!1,H(U),U=-1),A=!0;var le=w;try{t:{for(Q(F),y=u(b);y!==null&&!(y.expirationTime>F&&oe());){var je=y.callback;if(typeof je=="function"){y.callback=null,w=y.priorityLevel;var Ae=je(y.expirationTime<=F);if(F=o.unstable_now(),typeof Ae=="function"){y.callback=Ae,Q(F),ne=!0;break t}y===u(b)&&r(b),Q(F)}else r(b);y=u(b)}if(y!==null)ne=!0;else{var Qe=u(C);Qe!==null&&W(Y,Qe.startTime-F),ne=!1}}break e}finally{y=null,w=le,A=!1}ne=void 0}}finally{ne?re():S=!1}}}var re;if(typeof G=="function")re=function(){G(be)};else if(typeof MessageChannel<"u"){var J=new MessageChannel,ve=J.port2;J.port1.onmessage=be,re=function(){ve.postMessage(null)}}else re=function(){D(be,0)};function W(F,ne){U=D(function(){F(o.unstable_now())},ne)}o.unstable_IdlePriority=5,o.unstable_ImmediatePriority=1,o.unstable_LowPriority=4,o.unstable_NormalPriority=3,o.unstable_Profiling=null,o.unstable_UserBlockingPriority=2,o.unstable_cancelCallback=function(F){F.callback=null},o.unstable_forceFrameRate=function(F){0>F||125<F?console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"):Z=0<F?Math.floor(1e3/F):5},o.unstable_getCurrentPriorityLevel=function(){return w},o.unstable_next=function(F){switch(w){case 1:case 2:case 3:var ne=3;break;default:ne=w}var le=w;w=ne;try{return F()}finally{w=le}},o.unstable_requestPaint=function(){N=!0},o.unstable_runWithPriority=function(F,ne){switch(F){case 1:case 2:case 3:case 4:case 5:break;default:F=3}var le=w;w=F;try{return ne()}finally{w=le}},o.unstable_scheduleCallback=function(F,ne,le){var je=o.unstable_now();switch(typeof le=="object"&&le!==null?(le=le.delay,le=typeof le=="number"&&0<le?je+le:je):le=je,F){case 1:var Ae=-1;break;case 2:Ae=250;break;case 5:Ae=1073741823;break;case 4:Ae=1e4;break;default:Ae=5e3}return Ae=le+Ae,F={id:x++,callback:ne,priorityLevel:F,startTime:le,expirationTime:Ae,sortIndex:-1},le>je?(F.sortIndex=le,c(C,F),u(b)===null&&F===u(C)&&(_?(H(U),U=-1):_=!0,W(Y,le-je))):(F.sortIndex=Ae,c(b,F),E||A||(E=!0,S||(S=!0,re()))),F},o.unstable_shouldYield=oe,o.unstable_wrapCallback=function(F){var ne=w;return function(){var le=w;w=ne;try{return F.apply(this,arguments)}finally{w=le}}}})(vd)),vd}var Am;function $1(){return Am||(Am=1,gd.exports=P1()),gd.exports}var yd={exports:{}},yt={};var km;function ex(){if(km)return yt;km=1;var o=Dd();function c(x){var y="https://react.dev/errors/"+x;if(1<arguments.length){y+="?args[]="+encodeURIComponent(arguments[1]);for(var w=2;w<arguments.length;w++)y+="&args[]="+encodeURIComponent(arguments[w])}return"Minified React error #"+x+"; visit "+y+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}function u(){}var r={d:{f:u,r:function(){throw Error(c(522))},D:u,C:u,L:u,m:u,X:u,S:u,M:u},p:0,findDOMNode:null},h=Symbol.for("react.portal"),m=Symbol.for("react.recoverable"),g=Symbol.for("react.optimistic_key");function v(x,y,w){var A=3<arguments.length&&arguments[3]!==void 0?arguments[3]:null;return{$$typeof:h,key:A==null?null:A===g?g:""+A,children:x,containerInfo:y,implementation:w}}var b=o.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;function C(x,y){if(x==="font")return"";if(typeof y=="string")return y==="use-credentials"?y:""}return yt.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE=r,yt.browser=function(x){return{$$typeof:m,_reason:x}},yt.createPortal=function(x,y){var w=2<arguments.length&&arguments[2]!==void 0?arguments[2]:null;if(!y||y.nodeType!==1&&y.nodeType!==9&&y.nodeType!==11)throw Error(c(299));return v(x,y,null,w)},yt.flushSync=function(x){var y=b.T,w=r.p;try{if(b.T=null,r.p=2,x)return x()}finally{b.T=y,r.p=w,r.d.f()}},yt.preconnect=function(x,y){typeof x=="string"&&(y?(y=y.crossOrigin,y=typeof y=="string"?y==="use-credentials"?y:"":void 0):y=null,r.d.C(x,y))},yt.prefetchDNS=function(x){typeof x=="string"&&r.d.D(x)},yt.preinit=function(x,y){if(typeof x=="string"&&y&&typeof y.as=="string"){var w=y.as,A=C(w,y.crossOrigin),E=typeof y.integrity=="string"?y.integrity:void 0,_=typeof y.fetchPriority=="string"?y.fetchPriority:void 0;w==="style"?r.d.S(x,typeof y.precedence=="string"?y.precedence:void 0,{crossOrigin:A,integrity:E,fetchPriority:_}):w==="script"&&r.d.X(x,{crossOrigin:A,integrity:E,fetchPriority:_,nonce:typeof y.nonce=="string"?y.nonce:void 0})}},yt.preinitModule=function(x,y){if(typeof x=="string")if(typeof y=="object"&&y!==null){if(y.as==null||y.as==="script"){var w=C(y.as,y.crossOrigin);r.d.M(x,{crossOrigin:w,integrity:typeof y.integrity=="string"?y.integrity:void 0,nonce:typeof y.nonce=="string"?y.nonce:void 0,fetchPriority:typeof y.fetchPriority=="string"?y.fetchPriority:void 0})}}else y==null&&r.d.M(x)},yt.preload=function(x,y){if(typeof x=="string"&&typeof y=="object"&&y!==null&&typeof y.as=="string"){var w=y.as,A=C(w,y.crossOrigin);r.d.L(x,w,{crossOrigin:A,integrity:typeof y.integrity=="string"?y.integrity:void 0,nonce:typeof y.nonce=="string"?y.nonce:void 0,type:typeof y.type=="string"?y.type:void 0,fetchPriority:typeof y.fetchPriority=="string"?y.fetchPriority:void 0,referrerPolicy:typeof y.referrerPolicy=="string"?y.referrerPolicy:void 0,imageSrcSet:typeof y.imageSrcSet=="string"?y.imageSrcSet:void 0,imageSizes:typeof y.imageSizes=="string"?y.imageSizes:void 0,media:typeof y.media=="string"?y.media:void 0})}},yt.preloadModule=function(x,y){if(typeof x=="string")if(y){var w=C(y.as,y.crossOrigin);r.d.m(x,{as:typeof y.as=="string"&&y.as!=="script"?y.as:void 0,crossOrigin:w,integrity:typeof y.integrity=="string"?y.integrity:void 0,nonce:typeof y.nonce=="string"?y.nonce:void 0,fetchPriority:typeof y.fetchPriority=="string"?y.fetchPriority:void 0})}else r.d.m(x)},yt.requestFormReset=function(x){r.d.r(x)},yt.unstable_batchedUpdates=function(x,y){return x(y)},yt.useFormState=function(x,y,w){return b.H.useFormState(x,y,w)},yt.useFormStatus=function(){return b.H.useHostTransitionStatus()},yt.version="19.3.0",yt}var Cm;function tx(){if(Cm)return yd.exports;Cm=1;function o(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__>"u"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o)}catch(c){console.error(c)}}return o(),yd.exports=ex(),yd.exports}var Tm;function nx(){if(Tm)return Qi;Tm=1;var o=$1(),c=Dd(),u=tx();function r(e){var t="https://react.dev/errors/"+e;if(1<arguments.length){t+="?args[]="+encodeURIComponent(arguments[1]);for(var n=2;n<arguments.length;n++)t+="&args[]="+encodeURIComponent(arguments[n])}return"Minified React error #"+e+"; visit "+t+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}function h(e){return!(!e||e.nodeType!==1&&e.nodeType!==9&&e.nodeType!==11)}function m(e){for(var t=e,n=t;n&&!n.alternate;)t=n,(t.flags&4098)!==0&&(e=t.return),n=t.return;for(;t.return;)t=t.return;return t.tag===3?e:null}function g(e){if(e.tag===13){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function v(e){if(e.tag===31){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function b(e){if(m(e)!==e)throw Error(r(188))}function C(e){var t=e.alternate;if(!t){if(t=m(e),t===null)throw Error(r(188));return t!==e?null:e}for(var n=e,s=t;;){var i=n.return;if(i===null)break;var l=i.alternate;if(l===null){if(s=i.return,s!==null){n=s;continue}break}if(i.child===l.child){for(l=i.child;l;){if(l===n)return b(i),e;if(l===s)return b(i),t;l=l.sibling}throw Error(r(188))}if(n.return!==s.return)n=i,s=l;else{for(var d=!1,p=i.child;p;){if(p===n){d=!0,n=i,s=l;break}if(p===s){d=!0,s=i,n=l;break}p=p.sibling}if(!d){for(p=l.child;p;){if(p===n){d=!0,n=l,s=i;break}if(p===s){d=!0,s=l,n=i;break}p=p.sibling}if(!d)throw Error(r(189))}}if(n.alternate!==s)throw Error(r(190))}if(n.tag!==3)throw Error(r(188));return n.stateNode.current===n?e:t}function x(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e;for(e=e.child;e!==null;){if(t=x(e),t!==null)return t;e=e.sibling}return null}function y(e,t,n,s,i,l){for(;e!==null;){if((e.tag===5||e.tag===27||e.tag===6)&&n(e,s,i,l)||(e.tag!==22||e.memoizedState===null)&&(t||e.tag!==5&&e.tag!==27)&&y(e.child,t,n,s,i,l))return!0;e=e.sibling}return!1}function w(e){for(e=e.return;e!==null;){if(e.tag===3||e.tag===5||e.tag===27)return e;e=e.return}return null}function A(e){var t=!1;for(e=e.return;e!==null&&(e.tag===4&&(t=!0),!(e.tag===3||e.tag===5||e.tag===27));)e=e.return;return t}function E(e){var t=[null,null],n=w(e);return n===null||_(t,e,n.child,{foundSelf:!1}),t}function _(e,t,n,s){for(;n!==null;){if(n===t)s.foundSelf=!0;else if(n.tag===5||n.tag===27||n.tag===6){if(s.foundSelf)return e[1]=n,!0;e[0]=n}else if((n.tag!==22||n.memoizedState===null)&&_(e,t,n.child,s))return!0;n=n.sibling}return!1}function N(e){switch(e.tag){case 5:case 27:case 6:return e.stateNode;case 3:return e.stateNode.containerInfo;default:throw Error(r(559))}}var D=null,H=null;function G(e,t,n){return e===n?!0:e===t?(D=e,!0):!1}function Q(e,t,n){return e===n?(H=e,!1):e===t?(H!==null&&(D=e),!0):!1}function Y(e){if(e===null)return null;do e=e===null?null:e.return;while(e&&e.tag!==5&&e.tag!==27&&e.tag!==3);return e||null}function S(e,t,n){for(var s=0,i=e;i;i=n(i))s++;i=0;for(var l=t;l;l=n(l))i++;for(;0<s-i;)e=n(e),s--;for(;0<i-s;)t=n(t),i--;for(;s--;){if(e===t||t!==null&&e===t.alternate)return e;e=n(e),t=n(t)}return null}var U=Object.assign,Z=Symbol.for("react.element"),$=Symbol.for("react.transitional.element"),oe=Symbol.for("react.portal"),be=Symbol.for("react.fragment"),re=Symbol.for("react.strict_mode"),J=Symbol.for("react.profiler"),ve=Symbol.for("react.consumer"),W=Symbol.for("react.context"),F=Symbol.for("react.forward_ref"),ne=Symbol.for("react.suspense"),le=Symbol.for("react.suspense_list"),je=Symbol.for("react.memo"),Ae=Symbol.for("react.lazy"),Qe=Symbol.for("react.activity"),Ze=Symbol.for("react.legacy_hidden"),dt=Symbol.for("react.memo_cache_sentinel"),j=Symbol.for("react.view_transition"),I=Symbol.for("react.recoverable"),ee=Symbol.iterator;function ie(e){return e===null||typeof e!="object"?null:(e=ee&&e[ee]||e["@@iterator"],typeof e=="function"?e:null)}var fe=Symbol.for("react.client.reference");function ye(e){if(e==null)return null;if(typeof e=="function")return e.$$typeof===fe?null:e.displayName||e.name||null;if(typeof e=="string")return e;switch(e){case be:return"Fragment";case J:return"Profiler";case re:return"StrictMode";case ne:return"Suspense";case le:return"SuspenseList";case Qe:return"Activity";case j:return"ViewTransition"}if(typeof e=="object")switch(e.$$typeof){case oe:return"Portal";case W:return e.displayName||"Context";case ve:return(e._context.displayName||"Context")+".Consumer";case F:var t=e.render;return e=e.displayName,e||(e=t.displayName||t.name||"",e=e!==""?"ForwardRef("+e+")":"ForwardRef"),e;case je:return t=e.displayName||null,t!==null?t:ye(e.type)||"Memo";case Ae:t=e._payload,e=e._init;try{return ye(e(t))}catch{}}return null}var xe=Array.isArray,ae=c.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,he=u.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,ut={pending:!1,data:null,method:null,action:null},Qt=[],B=-1;function Re(e){return{current:e}}function Ie(e){0>B||(e.current=Qt[B],Qt[B]=null,B--)}function we(e,t){B++,Qt[B]=e.current,e.current=t}var st=Re(null),de=Re(null),Fe=Re(null),Ve=Re(null);function nn(e,t){switch(we(Fe,t),we(de,e),we(st,null),t.nodeType){case 9:case 11:e=(e=t.documentElement)&&(e=e.namespaceURI)?Sf(e):0;break;default:if(e=t.tagName,t=t.namespaceURI)t=Sf(t),e=Nf(t,e);else switch(e){case"svg":e=1;break;case"math":e=2;break;default:e=0}}Ie(st),we(st,e)}function Gn(){Ie(st),Ie(de),Ie(Fe)}function Js(e){var t=e.memoizedState;t!==null&&(Ls._currentValue=t.memoizedState,we(Ve,e)),t=st.current;var n=Nf(t,e.type);t!==n&&(we(de,e),we(st,n))}function ba(e){de.current===e&&(Ie(st),Ie(de)),Ve.current===e&&(Ie(Ve),Ls._currentValue=ut)}var Ws,Ja;function hn(e){if(Ws===void 0)try{throw Error()}catch(n){var t=n.stack.trim().match(/\n( *(at )?)/);Ws=t&&t[1]||"",Ja=-1<n.stack.indexOf(`
    at`)?" (<anonymous>)":-1<n.stack.indexOf("@")?"@unknown:0:0":""}return`
`+Ws+e+Ja}var K=!1;function me(e,t){if(!e||K)return"";K=!0;var n=Error.prepareStackTrace;Error.prepareStackTrace=void 0;try{var s={DetermineComponentFrameRoot:function(){try{if(t){var V=function(){throw Error()};if(Object.defineProperty(V.prototype,"props",{set:function(){throw Error()}}),typeof Reflect=="object"&&Reflect.construct){try{Reflect.construct(V,[])}catch(P){var O=P}Reflect.construct(e,[],V)}else{try{V.call()}catch(P){O=P}V=!1;try{var q=Object.getOwnPropertyDescriptor(e.prototype,"props");Object.defineProperty(e.prototype,"props",{configurable:!0,set:function(){throw Error()}}),V=!0,new e}finally{V&&(q!==void 0?Object.defineProperty(e.prototype,"props",q):delete e.prototype.props)}}}else{try{throw Error()}catch(P){O=P}(V=e())&&typeof V.catch=="function"&&V.catch(function(){})}}catch(P){if(P&&O&&typeof P.stack=="string")return[P.stack,O.stack]}return[null,null]}};s.DetermineComponentFrameRoot.displayName="DetermineComponentFrameRoot";var i=Object.getOwnPropertyDescriptor(s.DetermineComponentFrameRoot,"name");i&&i.configurable&&Object.defineProperty(s.DetermineComponentFrameRoot,"name",{value:"DetermineComponentFrameRoot"});var l=s.DetermineComponentFrameRoot(),d=l[0],p=l[1];if(d&&p){var k=d.split(`
`),R=p.split(`
`);for(i=s=0;s<k.length&&!k[s].includes("DetermineComponentFrameRoot");)s++;for(;i<R.length&&!R[i].includes("DetermineComponentFrameRoot");)i++;if(s===k.length||i===R.length)for(s=k.length-1,i=R.length-1;1<=s&&0<=i&&k[s]!==R[i];)i--;for(;1<=s&&0<=i;s--,i--)if(k[s]!==R[i]){if(s!==1||i!==1)do if(s--,i--,0>i||k[s]!==R[i]){var L=`
`+k[s].replace(" at new "," at ");return e.displayName&&L.includes("<anonymous>")&&(L=L.replace("<anonymous>",e.displayName)),L}while(1<=s&&0<=i);break}}}finally{K=!1,Error.prepareStackTrace=n}return(n=e?e.displayName||e.name:"")?hn(n):""}function Be(e,t){switch(e.tag){case 26:case 27:case 5:return hn(e.type);case 16:return hn("Lazy");case 13:return e.child!==t&&t!==null?hn("Suspense Fallback"):hn("Suspense");case 19:return hn("SuspenseList");case 0:case 15:return me(e.type,!1);case 11:return me(e.type.render,!1);case 1:return me(e.type,!0);case 31:return hn("Activity");case 30:return hn("ViewTransition");default:return""}}function ja(e){try{var t="",n=null;do t+=Be(e,n),n=e,e=e.return;while(e);return t}catch(s){return`
Error generating stack: `+s.message+`
`+s.stack}}var Bn=Object.prototype.hasOwnProperty,Ul=o.unstable_scheduleCallback,Hl=o.unstable_cancelCallback,$v=o.unstable_shouldYield,ey=o.unstable_requestPaint,Mt=o.unstable_now,ty=o.unstable_getCurrentPriorityLevel,Fd=o.unstable_ImmediatePriority,Jd=o.unstable_UserBlockingPriority,to=o.unstable_NormalPriority,ny=o.unstable_LowPriority,Wd=o.unstable_IdlePriority,ay=o.log,sy=o.unstable_setDisableYieldValue,Ps=null,qt=null;function Yn(e){if(typeof ay=="function"&&sy(e),qt&&typeof qt.setStrictMode=="function")try{qt.setStrictMode(Ps,e)}catch{}}var Dt=Math.clz32?Math.clz32:ly,iy=Math.log,oy=Math.LN2;function ly(e){return e>>>=0,e===0?32:31-(iy(e)/oy|0)|0}var no=256,ao=262144,so=4194304;function Aa(e){var t=e&42;if(t!==0)return t;switch(e&-e){case 1:return 1;case 2:return 2;case 4:return 4;case 8:return 8;case 16:return 16;case 32:return 32;case 64:return 64;case 128:return 128;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:return e&-e;case 262144:case 524288:case 1048576:case 2097152:return e&3932160;case 4194304:case 8388608:case 16777216:case 33554432:return e&62914560;case 67108864:return 67108864;case 134217728:return 134217728;case 268435456:return 268435456;case 536870912:return 536870912;case 1073741824:return 0;default:return e}}function io(e,t,n){var s=e.pendingLanes;if(s===0)return 0;var i=0,l=e.suspendedLanes,d=e.pingedLanes;e=e.warmLanes;var p=s&134217727;return p!==0?(s=p&~l,s!==0?i=Aa(s):(d&=p,d!==0?i=Aa(d):n||(n=p&~e,n!==0&&(i=Aa(n))))):(p=s&~l,p!==0?i=Aa(p):d!==0?i=Aa(d):n||(n=s&~e,n!==0&&(i=Aa(n)))),i===0?0:t!==0&&t!==i&&(t&l)===0&&(l=i&-i,n=t&-t,l>=n||l===32&&(n&4194048)!==0)?t:i}function $s(e,t){return(e.pendingLanes&~(e.suspendedLanes&~e.pingedLanes)&t)===0}function Pd(e,t){(t&8)!==0&&(t|=t&32);var n=e.entangledLanes;if(n!==0)for(e=e.entanglements,n&=t;0<n;){var s=31-Dt(n),i=1<<s;t|=e[s],n&=~i}return t}function ry(e,t){switch(e){case 1:case 2:case 4:case 8:case 64:return t+250;case 16:case 32:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return t+5e3;case 4194304:case 8388608:case 16777216:case 33554432:return-1;case 67108864:case 134217728:case 268435456:case 536870912:case 1073741824:return-1;default:return-1}}function $d(){var e=so;return so<<=1,(so&62914560)===0&&(so=4194304),e}function Il(e){for(var t=[],n=0;31>n;n++)t.push(e);return t}function ei(e,t){e.pendingLanes|=t,t!==268435456&&(e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0)}function cy(e,t,n,s,i,l){var d=e.pendingLanes;e.pendingLanes=n,e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0,e.expiredLanes&=n,e.entangledLanes&=n,e.errorRecoveryDisabledLanes&=n,e.shellSuspendCounter=0;var p=e.entanglements,k=e.expirationTimes,R=e.hiddenUpdates;for(n=d&~n;0<n;){var L=31-Dt(n),V=1<<L;p[L]=0,k[L]=-1;var O=R[L];if(O!==null)for(R[L]=null,L=0;L<O.length;L++){var q=O[L];q!==null&&(q.lane&=-536870913)}n&=~V}s!==0&&eu(e,s,0),l!==0&&i===0&&e.tag!==0&&(e.suspendedLanes|=l&~(d&~t))}function eu(e,t,n){e.pendingLanes|=t,e.suspendedLanes&=~t;var s=31-Dt(t);e.entangledLanes|=t,e.entanglements[s]=e.entanglements[s]|1073741824|n&261930}function tu(e,t){var n=e.entangledLanes|=t;for(e=e.entanglements;n;){var s=31-Dt(n),i=1<<s;i&t|e[s]&t&&(e[s]|=t),n&=~i}}function nu(e,t){var n=t&-t;return n=(n&42)!==0?1:Gl(n),(n&(e.suspendedLanes|t))!==0?0:n}function Gl(e){switch(e){case 2:e=1;break;case 8:e=4;break;case 32:e=16;break;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:case 4194304:case 8388608:case 16777216:case 33554432:e=128;break;case 268435456:e=134217728;break;default:e=0}return e}function Bl(e){return e&=-e,2<e?8<e?(e&134217727)!==0?32:268435456:8:2}function au(){var e=he.p;return e!==0?e:(e=window.event,e===void 0?32:um(e.type))}function su(e,t){var n=he.p;try{return he.p=e,t()}finally{he.p=n}}var Cn=Math.random().toString(36).slice(2),ht="__reactFiber$"+Cn,Tt="__reactProps$"+Cn,Wa="__reactContainer$"+Cn,iu="__reactEvents$"+Cn,dy="__reactListeners$"+Cn,uy="__reactHandles$"+Cn,ou="__reactResources$"+Cn,ti="__reactMarker$"+Cn,oo="__reactLoad$"+Cn;function lo(e){delete e[ht],delete e[Tt],delete e[dy],delete e[uy]}function ka(e){var t;if(t=e[ht])return t;for(var n=e.parentNode;n;){if(t=n[Wa]||n[ht]){if(n=t.alternate,t.child!==null||n!==null&&n.child!==null)for(e=Vf(e);e!==null;){if(n=e[ht])return n;e=Vf(e)}return t}e=n,n=e.parentNode}return null}function Pa(e){if(e=e[ht]||e[Wa]){var t=e.tag;if(t===5||t===6||t===13||t===31||t===26||t===27||t===3)return e}return null}function ni(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e.stateNode;throw Error(r(33))}function $a(e){var t=e[ou];return t||(t=e[ou]={hoistableStyles:new Map,hoistableScripts:new Map}),t}function ot(e){e[ti]=!0}function lu(e){e[oo]=void 0}var ru=new Set,cu={};function Ca(e,t){es(e,t),es(e+"Capture",t)}function es(e,t){for(cu[e]=t,e=0;e<t.length;e++)ru.add(t[e])}var hy=RegExp("^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"),du={},uu={};function py(e){return Bn.call(uu,e)?!0:Bn.call(du,e)?!1:hy.test(e)?uu[e]=!0:(du[e]=!0,!1)}var ze=!1;function hu(){var e=ze;return ze=!1,e}function ro(e,t,n){if(py(t))if(n===null)e.removeAttribute(t);else{switch(typeof n){case"undefined":case"function":case"symbol":e.removeAttribute(t);return;case"boolean":var s=t.toLowerCase().slice(0,5);if(s!=="data-"&&s!=="aria-"){e.removeAttribute(t);return}}e.setAttribute(t,n)}}function co(e,t,n){if(n===null)e.removeAttribute(t);else{switch(typeof n){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(t);return}e.setAttribute(t,n)}}function Tn(e,t,n,s){if(s===null)e.removeAttribute(n);else{switch(typeof s){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(n);return}e.setAttributeNS(t,n,s)}}function Lt(e){switch(typeof e){case"bigint":case"boolean":case"number":case"string":case"undefined":return e;case"object":return e;default:return""}}function pu(e){var t=e.type;return(e=e.nodeName)&&e.toLowerCase()==="input"&&(t==="checkbox"||t==="radio")}function fy(e,t,n){var s=Object.getOwnPropertyDescriptor(e.constructor.prototype,t);if(!e.hasOwnProperty(t)&&typeof s<"u"&&typeof s.get=="function"&&typeof s.set=="function"){var i=s.get,l=s.set;return Object.defineProperty(e,t,{configurable:!0,get:function(){return i.call(this)},set:function(d){n=""+d,l.call(this,d)}}),Object.defineProperty(e,t,{enumerable:s.enumerable}),{getValue:function(){return n},setValue:function(d){n=""+d},stopTracking:function(){e._valueTracker=null,delete e[t]}}}}function Yl(e){if(!e._valueTracker){var t=pu(e)?"checked":"value";e._valueTracker=fy(e,t,""+e[t])}}function fu(e){if(!e)return!1;var t=e._valueTracker;if(!t)return!0;var n=t.getValue(),s="";return e&&(s=pu(e)?e.checked?"true":"false":e.value),e=s,e!==n?(t.setValue(e),!0):!1}var my=/[\n"\\]/g;function Zt(e){return e.replace(my,function(t){return"\\"+t.charCodeAt(0).toString(16)+" "})}function Xl(e,t,n,s,i,l,d,p){e.name="",d!=null&&typeof d!="function"&&typeof d!="symbol"&&typeof d!="boolean"?e.type=d:e.removeAttribute("type"),t!=null?d==="number"?(t===0&&e.value===""||e.value!=t)&&(e.value=""+Lt(t)):e.value!==""+Lt(t)&&(e.value=""+Lt(t)):d!=="submit"&&d!=="reset"||e.removeAttribute("value"),t!=null?d==="number"&&e.value==t?Vl(e,Lt(e.value)):Vl(e,Lt(t)):n!=null?Vl(e,Lt(n)):s!=null&&e.removeAttribute("value"),i==null&&l!=null&&(e.defaultChecked=!!l),i!=null&&(e.checked=i&&typeof i!="function"&&typeof i!="symbol"),p!=null&&typeof p!="function"&&typeof p!="symbol"&&typeof p!="boolean"?e.name=""+Lt(p):e.removeAttribute("name")}function mu(e,t,n,s,i,l,d,p){if(l!=null&&typeof l!="function"&&typeof l!="symbol"&&typeof l!="boolean"&&(e.type=l),t!=null||n!=null){if(!(l!=="submit"&&l!=="reset"||t!=null)){Yl(e);return}n=n!=null?""+Lt(n):"",t=t!=null?""+Lt(t):n,p||t===e.value||(e.value=t),e.defaultValue=t}s=s??i,s=typeof s!="function"&&typeof s!="symbol"&&!!s,e.checked=p?e.checked:!!s,e.defaultChecked=!!s,d!=null&&typeof d!="function"&&typeof d!="symbol"&&typeof d!="boolean"&&(e.name=d),Yl(e)}function Vl(e,t){e.defaultValue!==""+t&&(e.defaultValue=""+t)}function ts(e,t,n,s){if(e=e.options,t){t={};for(var i=0;i<n.length;i++)t["$"+n[i]]=!0;for(n=0;n<e.length;n++)i=t.hasOwnProperty("$"+e[n].value),e[n].selected!==i&&(e[n].selected=i),i&&s&&(e[n].defaultSelected=!0)}else{for(n=""+Lt(n),t=null,i=0;i<e.length;i++){if(e[i].value===n){e[i].selected=!0,s&&(e[i].defaultSelected=!0);return}t!==null||e[i].disabled||(t=e[i])}t!==null&&(t.selected=!0)}}function gu(e,t,n){if(t!=null&&(t=""+Lt(t),t!==e.value&&(e.value=t),n==null)){e.defaultValue!==t&&(e.defaultValue=t);return}e.defaultValue=n!=null?""+Lt(n):""}function vu(e,t,n,s){if(t==null){if(s!=null){if(n!=null)throw Error(r(92));if(xe(s)){if(1<s.length)throw Error(r(93));s=s[0]}n=s}n==null&&(n=""),t=n}n=Lt(t),e.defaultValue=n,s=e.textContent,s===n&&s!==""&&s!==null&&(e.value=s),Yl(e)}function ns(e,t){if(t){var n=e.firstChild;if(n&&n===e.lastChild&&n.nodeType===3){n.nodeValue=t;return}}e.textContent=t}var gy=new Set("animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(" "));function yu(e,t,n){var s=t.indexOf("--")===0;n==null||typeof n=="boolean"||n===""?s?e.setProperty(t,""):t==="float"?e.cssFloat="":e[t]="":s?e.setProperty(t,n):typeof n!="number"||n===0||gy.has(t)?t==="float"?e.cssFloat=n:e[t]=(""+n).trim():e[t]=n+"px"}function xu(e,t,n){if(t!=null&&typeof t!="object")throw Error(r(62));if(e=e.style,n!=null){for(var s in n)!n.hasOwnProperty(s)||t!=null&&t.hasOwnProperty(s)||(s.indexOf("--")===0?e.setProperty(s,""):s==="float"?e.cssFloat="":e[s]="",ze=!0);for(var i in t)s=t[i],t.hasOwnProperty(i)&&n[i]!==s&&(yu(e,i,s),ze=!0)}else for(var l in t)t.hasOwnProperty(l)&&yu(e,l,t[l])}function Ql(e){if(e.indexOf("-")===-1)return!1;switch(e){case"annotation-xml":case"color-profile":case"font-face":case"font-face-src":case"font-face-uri":case"font-face-format":case"font-face-name":case"missing-glyph":return!1;default:return!0}}var vy=new Map([["acceptCharset","accept-charset"],["htmlFor","for"],["httpEquiv","http-equiv"],["crossOrigin","crossorigin"],["accentHeight","accent-height"],["alignmentBaseline","alignment-baseline"],["arabicForm","arabic-form"],["baselineShift","baseline-shift"],["capHeight","cap-height"],["clipPath","clip-path"],["clipRule","clip-rule"],["colorInterpolation","color-interpolation"],["colorInterpolationFilters","color-interpolation-filters"],["colorProfile","color-profile"],["colorRendering","color-rendering"],["dominantBaseline","dominant-baseline"],["enableBackground","enable-background"],["fillOpacity","fill-opacity"],["fillRule","fill-rule"],["floodColor","flood-color"],["floodOpacity","flood-opacity"],["fontFamily","font-family"],["fontSize","font-size"],["fontSizeAdjust","font-size-adjust"],["fontStretch","font-stretch"],["fontStyle","font-style"],["fontVariant","font-variant"],["fontWeight","font-weight"],["glyphName","glyph-name"],["glyphOrientationHorizontal","glyph-orientation-horizontal"],["glyphOrientationVertical","glyph-orientation-vertical"],["horizAdvX","horiz-adv-x"],["horizOriginX","horiz-origin-x"],["imageRendering","image-rendering"],["letterSpacing","letter-spacing"],["lightingColor","lighting-color"],["markerEnd","marker-end"],["markerMid","marker-mid"],["markerStart","marker-start"],["maskType","mask-type"],["overlinePosition","overline-position"],["overlineThickness","overline-thickness"],["paintOrder","paint-order"],["panose-1","panose-1"],["pointerEvents","pointer-events"],["renderingIntent","rendering-intent"],["shapeRendering","shape-rendering"],["stopColor","stop-color"],["stopOpacity","stop-opacity"],["strikethroughPosition","strikethrough-position"],["strikethroughThickness","strikethrough-thickness"],["strokeDasharray","stroke-dasharray"],["strokeDashoffset","stroke-dashoffset"],["strokeLinecap","stroke-linecap"],["strokeLinejoin","stroke-linejoin"],["strokeMiterlimit","stroke-miterlimit"],["strokeOpacity","stroke-opacity"],["strokeWidth","stroke-width"],["textAnchor","text-anchor"],["textDecoration","text-decoration"],["textRendering","text-rendering"],["transformOrigin","transform-origin"],["underlinePosition","underline-position"],["underlineThickness","underline-thickness"],["unicodeBidi","unicode-bidi"],["unicodeRange","unicode-range"],["unitsPerEm","units-per-em"],["vAlphabetic","v-alphabetic"],["vHanging","v-hanging"],["vIdeographic","v-ideographic"],["vMathematical","v-mathematical"],["vectorEffect","vector-effect"],["vertAdvY","vert-adv-y"],["vertOriginX","vert-origin-x"],["vertOriginY","vert-origin-y"],["wordSpacing","word-spacing"],["writingMode","writing-mode"],["xmlnsXlink","xmlns:xlink"],["xHeight","x-height"]]),yy=/^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;function uo(e){return yy.test(""+e)?"javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')":e}function pn(){}var Zl=null;function Kl(e){return e=e.target||e.srcElement||window,e.correspondingUseElement&&(e=e.correspondingUseElement),e.nodeType===3?e.parentNode:e}var as=null,ss=null;function wu(e){var t=Pa(e);if(t&&(e=t.stateNode)){var n=e[Tt]||null;e:switch(e=t.stateNode,t.type){case"input":if(Xl(e,n.value,n.defaultValue,n.defaultValue,n.checked,n.defaultChecked,n.type,n.name),t=n.name,n.type==="radio"&&t!=null){for(n=e;n.parentNode;)n=n.parentNode;for(n=n.querySelectorAll('input[name="'+Zt(""+t)+'"][type="radio"]'),t=0;t<n.length;t++){var s=n[t];if(s!==e&&s.form===e.form){var i=s[Tt]||null;if(!i)throw Error(r(90));Xl(s,i.value,i.defaultValue,i.defaultValue,i.checked,i.defaultChecked,i.type,i.name)}}for(t=0;t<n.length;t++)s=n[t],s.form===e.form&&fu(s)}break e;case"textarea":gu(e,n.value,n.defaultValue);break e;case"select":t=n.value,t!=null&&ts(e,!!n.multiple,t,!1)}}}var Fl=!1;function bu(e,t,n){if(Fl)return e(t,n);Fl=!0;try{var s=e(t);return s}finally{if(Fl=!1,(as!==null||ss!==null)&&(ul(),as&&(t=as,e=ss,ss=as=null,wu(t),e)))for(t=0;t<e.length;t++)wu(e[t])}}function ai(e,t){var n=e.stateNode;if(n===null)return null;var s=n[Tt]||null;if(s===null)return null;n=s[t];e:switch(t){case"onClick":case"onClickCapture":case"onDoubleClick":case"onDoubleClickCapture":case"onMouseDown":case"onMouseDownCapture":case"onMouseMove":case"onMouseMoveCapture":case"onMouseUp":case"onMouseUpCapture":case"onMouseEnter":(s=!s.disabled)||(e=e.type,s=!(e==="button"||e==="input"||e==="select"||e==="textarea")),e=!s;break e;default:e=!1}if(e)return null;if(n&&typeof n!="function")throw Error(r(231,t,typeof n));return n}var Sn=!(typeof window>"u"||typeof window.document>"u"||typeof window.document.createElement>"u"),Jl=!1;if(Sn)try{var si={};Object.defineProperty(si,"passive",{get:function(){Jl=!0}}),window.addEventListener("test",si,si),window.removeEventListener("test",si,si)}catch{Jl=!1}var Xn=null,Wl=null,ho=null;function ju(){if(ho)return ho;var e,t=Wl,n=t.length,s,i="value"in Xn?Xn.value:Xn.textContent,l=i.length;for(e=0;e<n&&t[e]===i[e];e++);var d=n-e;for(s=1;s<=d&&t[n-s]===i[l-s];s++);return ho=i.slice(e,1<s?1-s:void 0)}function po(e){var t=e.keyCode;return"charCode"in e?(e=e.charCode,e===0&&t===13&&(e=13)):e=t,e===10&&(e=13),32<=e||e===13?e:0}function fo(){return!0}function Au(){return!1}function bt(e){function t(n,s,i,l,d){this._reactName=n,this._targetInst=i,this.type=s,this.nativeEvent=l,this.target=d,this.currentTarget=null;for(var p in e)e.hasOwnProperty(p)&&(n=e[p],this[p]=n?n(l):l[p]);return this.isDefaultPrevented=(l.defaultPrevented!=null?l.defaultPrevented:l.returnValue===!1)?fo:Au,this.isPropagationStopped=Au,this}return U(t.prototype,{preventDefault:function(){this.defaultPrevented=!0;var n=this.nativeEvent;n&&(n.preventDefault?n.preventDefault():typeof n.returnValue!="unknown"&&(n.returnValue=!1),this.isDefaultPrevented=fo)},stopPropagation:function(){var n=this.nativeEvent;n&&(n.stopPropagation?n.stopPropagation():typeof n.cancelBubble!="unknown"&&(n.cancelBubble=!0),this.isPropagationStopped=fo)},persist:function(){},isPersistent:fo}),t}var Vn={eventPhase:0,bubbles:0,cancelable:0,timeStamp:function(e){return e.timeStamp||Date.now()},defaultPrevented:0,isTrusted:0},mo=bt(Vn),ii=U({},Vn,{view:0,detail:0}),xy=bt(ii),Pl,$l,oi,go=U({},ii,{screenX:0,screenY:0,clientX:0,clientY:0,pageX:0,pageY:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,getModifierState:tr,button:0,buttons:0,relatedTarget:function(e){return e.relatedTarget===void 0?e.fromElement===e.srcElement?e.toElement:e.fromElement:e.relatedTarget},movementX:function(e){return"movementX"in e?e.movementX:(e!==oi&&(oi&&e.type==="mousemove"?(Pl=e.screenX-oi.screenX,$l=e.screenY-oi.screenY):$l=Pl=0,oi=e),Pl)},movementY:function(e){return"movementY"in e?e.movementY:$l}}),ku=bt(go),wy=U({},go,{dataTransfer:0}),by=bt(wy),jy=U({},ii,{relatedTarget:0}),er=bt(jy),Ay=U({},Vn,{animationName:0,elapsedTime:0,pseudoElement:0}),ky=bt(Ay),Cy=U({},Vn,{clipboardData:function(e){return"clipboardData"in e?e.clipboardData:window.clipboardData}}),Ty=bt(Cy),Sy=U({},Vn,{data:0}),Cu=bt(Sy),Ny={Esc:"Escape",Spacebar:" ",Left:"ArrowLeft",Up:"ArrowUp",Right:"ArrowRight",Down:"ArrowDown",Del:"Delete",Win:"OS",Menu:"ContextMenu",Apps:"ContextMenu",Scroll:"ScrollLock",MozPrintableKey:"Unidentified"},_y={8:"Backspace",9:"Tab",12:"Clear",13:"Enter",16:"Shift",17:"Control",18:"Alt",19:"Pause",20:"CapsLock",27:"Escape",32:" ",33:"PageUp",34:"PageDown",35:"End",36:"Home",37:"ArrowLeft",38:"ArrowUp",39:"ArrowRight",40:"ArrowDown",45:"Insert",46:"Delete",112:"F1",113:"F2",114:"F3",115:"F4",116:"F5",117:"F6",118:"F7",119:"F8",120:"F9",121:"F10",122:"F11",123:"F12",144:"NumLock",145:"ScrollLock",224:"Meta"},Ey={Alt:"altKey",Control:"ctrlKey",Meta:"metaKey",Shift:"shiftKey"};function Oy(e){var t=this.nativeEvent;return t.getModifierState?t.getModifierState(e):(e=Ey[e])?!!t[e]:!1}function tr(){return Oy}var zy=U({},ii,{key:function(e){if(e.key){var t=Ny[e.key]||e.key;if(t!=="Unidentified")return t}return e.type==="keypress"?(e=po(e),e===13?"Enter":String.fromCharCode(e)):e.type==="keydown"||e.type==="keyup"?_y[e.keyCode]||"Unidentified":""},code:0,location:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,repeat:0,locale:0,getModifierState:tr,charCode:function(e){return e.type==="keypress"?po(e):0},keyCode:function(e){return e.type==="keydown"||e.type==="keyup"?e.keyCode:0},which:function(e){return e.type==="keypress"?po(e):e.type==="keydown"||e.type==="keyup"?e.keyCode:0}}),Ry=bt(zy),My=U({},go,{pointerId:0,width:0,height:0,pressure:0,tangentialPressure:0,tiltX:0,tiltY:0,twist:0,pointerType:0,isPrimary:0}),Tu=bt(My),qy=U({},Vn,{submitter:0}),Dy=bt(qy),Ly=U({},ii,{touches:0,targetTouches:0,changedTouches:0,altKey:0,metaKey:0,ctrlKey:0,shiftKey:0,getModifierState:tr}),Uy=bt(Ly),Hy=U({},Vn,{propertyName:0,elapsedTime:0,pseudoElement:0}),Iy=bt(Hy),Gy=U({},go,{deltaX:function(e){return"deltaX"in e?e.deltaX:"wheelDeltaX"in e?-e.wheelDeltaX:0},deltaY:function(e){return"deltaY"in e?e.deltaY:"wheelDeltaY"in e?-e.wheelDeltaY:"wheelDelta"in e?-e.wheelDelta:0},deltaZ:0,deltaMode:0}),By=bt(Gy),Yy=U({},Vn,{newState:0,oldState:0,source:0}),Xy=bt(Yy),Vy=[9,13,27,32],nr=Sn&&"CompositionEvent"in window,li=null;Sn&&"documentMode"in document&&(li=document.documentMode);var Qy=Sn&&"TextEvent"in window&&!li,Su=Sn&&(!nr||li&&8<li&&11>=li),Nu=" ",_u=!1;function Eu(e,t){switch(e){case"keyup":return Vy.indexOf(t.keyCode)!==-1;case"keydown":return t.keyCode!==229;case"keypress":case"mousedown":case"focusout":return!0;default:return!1}}function Ou(e){return e=e.detail,typeof e=="object"&&"data"in e?e.data:null}var is=!1;function Zy(e,t){switch(e){case"compositionend":return Ou(t);case"keypress":return t.which!==32?null:(_u=!0,Nu);case"textInput":return e=t.data,e===Nu&&_u?null:e;default:return null}}function Ky(e,t){if(is)return e==="compositionend"||!nr&&Eu(e,t)?(e=ju(),ho=Wl=Xn=null,is=!1,e):null;switch(e){case"paste":return null;case"keypress":if(!(t.ctrlKey||t.altKey||t.metaKey)||t.ctrlKey&&t.altKey){if(t.char&&1<t.char.length)return t.char;if(t.which)return String.fromCharCode(t.which)}return null;case"compositionend":return Su&&t.locale!=="ko"?null:t.data;default:return null}}var Fy={color:!0,date:!0,datetime:!0,"datetime-local":!0,email:!0,month:!0,number:!0,password:!0,range:!0,search:!0,tel:!0,text:!0,time:!0,url:!0,week:!0};function zu(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t==="input"?!!Fy[e.type]:t==="textarea"}function Ru(e,t,n,s){as?ss?ss.push(s):ss=[s]:as=s,t=vl(t,"onChange"),0<t.length&&(n=new mo("onChange","change",null,n,s),e.push({event:n,listeners:t}))}var ri=null,ci=null;function Jy(e){bf(e,0)}function vo(e){var t=ni(e);if(fu(t))return e}function Mu(e,t){if(e==="change")return t}var qu=!1;if(Sn){var ar;if(Sn){var sr="oninput"in document;if(!sr){var Du=document.createElement("div");Du.setAttribute("oninput","return;"),sr=typeof Du.oninput=="function"}ar=sr}else ar=!1;qu=ar&&(!document.documentMode||9<document.documentMode)}function Lu(){ri&&(ri.detachEvent("onpropertychange",Uu),ci=ri=null)}function Uu(e){if(e.propertyName==="value"&&vo(ci)){var t=[];Ru(t,ci,e,Kl(e)),bu(Jy,t)}}function Wy(e,t,n){e==="focusin"?(Lu(),ri=t,ci=n,ri.attachEvent("onpropertychange",Uu)):e==="focusout"&&Lu()}function Py(e){if(e==="selectionchange"||e==="keyup"||e==="keydown")return vo(ci)}function $y(e,t){if(e==="click")return vo(t)}function e0(e,t){if(e==="input"||e==="change")return vo(t)}function t0(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var Ut=typeof Object.is=="function"?Object.is:t0;function di(e,t){if(Ut(e,t))return!0;if(typeof e!="object"||e===null||typeof t!="object"||t===null)return!1;var n=Object.keys(e),s=Object.keys(t);if(n.length!==s.length)return!1;for(s=0;s<n.length;s++){var i=n[s];if(!Bn.call(t,i)||!Ut(e[i],t[i]))return!1}return!0}function ir(e){if(e=e||(typeof document<"u"?document:void 0),typeof e>"u")return null;try{return e.activeElement||e.body}catch{return e.body}}function Hu(e){for(;e&&e.firstChild;)e=e.firstChild;return e}function Iu(e,t){var n=Hu(e);e=0;for(var s;n;){if(n.nodeType===3){if(s=e+n.textContent.length,e<=t&&s>=t)return{node:n,offset:t-e};e=s}e:{for(;n;){if(n.nextSibling){n=n.nextSibling;break e}n=n.parentNode}n=void 0}n=Hu(n)}}function Gu(e,t){return e&&t?e===t?!0:e&&e.nodeType===3?!1:t&&t.nodeType===3?Gu(e,t.parentNode):"contains"in e?e.contains(t):e.compareDocumentPosition?!!(e.compareDocumentPosition(t)&16):!1:!1}function Bu(e){e=e!=null&&e.ownerDocument!=null&&e.ownerDocument.defaultView!=null?e.ownerDocument.defaultView:window;for(var t=ir(e.document);t instanceof e.HTMLIFrameElement;){try{var n=typeof t.contentWindow.location.href=="string"}catch{n=!1}if(n)e=t.contentWindow;else break;t=ir(e.document)}return t}function or(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t&&(t==="input"&&(e.type==="text"||e.type==="search"||e.type==="tel"||e.type==="url"||e.type==="password")||t==="textarea"||e.contentEditable==="true")}var n0=Sn&&"documentMode"in document&&11>=document.documentMode,os=null,lr=null,ui=null,rr=!1;function Yu(e,t,n){var s=n.window===n?n.document:n.nodeType===9?n:n.ownerDocument;rr||os==null||os!==ir(s)||(s=os,"selectionStart"in s&&or(s)?s={start:s.selectionStart,end:s.selectionEnd}:(s=(s.ownerDocument&&s.ownerDocument.defaultView||window).getSelection(),s={anchorNode:s.anchorNode,anchorOffset:s.anchorOffset,focusNode:s.focusNode,focusOffset:s.focusOffset}),ui&&di(ui,s)||(ui=s,s=vl(lr,"onSelect"),0<s.length&&(t=new mo("onSelect","select",null,t,n),e.push({event:t,listeners:s}),t.target=os)))}function Ta(e,t){var n={};return n[e.toLowerCase()]=t.toLowerCase(),n["Webkit"+e]="webkit"+t,n["Moz"+e]="moz"+t,n}var ls={animationend:Ta("Animation","AnimationEnd"),animationiteration:Ta("Animation","AnimationIteration"),animationstart:Ta("Animation","AnimationStart"),transitionrun:Ta("Transition","TransitionRun"),transitionstart:Ta("Transition","TransitionStart"),transitioncancel:Ta("Transition","TransitionCancel"),transitionend:Ta("Transition","TransitionEnd")},cr={},Xu={};Sn&&(Xu=document.createElement("div").style,"AnimationEvent"in window||(delete ls.animationend.animation,delete ls.animationiteration.animation,delete ls.animationstart.animation),"TransitionEvent"in window||delete ls.transitionend.transition);function Sa(e){if(cr[e])return cr[e];if(!ls[e])return e;var t=ls[e],n;for(n in t)if(t.hasOwnProperty(n)&&n in Xu)return cr[e]=t[n];return e}var Vu=Sa("animationend"),Qu=Sa("animationiteration"),Zu=Sa("animationstart"),a0=Sa("transitionrun"),s0=Sa("transitionstart"),i0=Sa("transitioncancel"),Ku=Sa("transitionend"),Fu=new Map,dr="abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error fullscreenChange fullscreenError gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");dr.push("scrollEnd");function an(e,t){Fu.set(e,t),Ca(t,[e])}var o0=0;function Nn(e,t){if(e.name!=null&&e.name!=="auto")return e.name;if(t.autoName!==null)return t.autoName;e=rn.identifierPrefix;var n=o0++;return e="_"+e+"t_"+n.toString(32)+"_",t.autoName=e}function Ju(e){if(e==null||typeof e=="string")return e;var t=null,n=Ss;if(n!==null)for(var s=0;s<n.length;s++){var i=e[n[s]];if(i!=null){if(i==="none")return"none";t=t==null?i:t+(" "+i)}}return t??e.default}function _n(e,t){return e=Ju(e),t=Ju(t),t==null?e==="auto"?null:e:t==="auto"?null:t}var yo=typeof reportError=="function"?reportError:function(e){if(typeof window=="object"&&typeof window.ErrorEvent=="function"){var t=new window.ErrorEvent("error",{bubbles:!0,cancelable:!0,message:typeof e=="object"&&e!==null&&typeof e.message=="string"?String(e.message):String(e),error:e});if(!window.dispatchEvent(t))return}else if(typeof process=="object"&&typeof process.emit=="function"){process.emit("uncaughtException",e);return}console.error(e)},Kt=[],rs=0,ur=0;function xo(){for(var e=rs,t=ur=rs=0;t<e;){var n=Kt[t];Kt[t++]=null;var s=Kt[t];Kt[t++]=null;var i=Kt[t];Kt[t++]=null;var l=Kt[t];if(Kt[t++]=null,s!==null&&i!==null){var d=s.pending;d===null?i.next=i:(i.next=d.next,d.next=i),s.pending=i}l!==0&&Wu(n,i,l)}}function wo(e,t,n,s){Kt[rs++]=e,Kt[rs++]=t,Kt[rs++]=n,Kt[rs++]=s,ur|=s,e.lanes|=s,e=e.alternate,e!==null&&(e.lanes|=s)}function hr(e,t,n,s){return wo(e,t,n,s),bo(e)}function Na(e,t){return wo(e,null,null,t),bo(e)}function Wu(e,t,n){e.lanes|=n;var s=e.alternate;s!==null&&(s.lanes|=n);for(var i=!1,l=e.return;l!==null;)l.childLanes|=n,s=l.alternate,s!==null&&(s.childLanes|=n),l.tag===22&&(e=l.stateNode,e===null||e._visibility&1||(i=!0)),e=l,l=l.return;return e.tag===3?(l=e.stateNode,i&&t!==null&&(i=31-Dt(n),e=l.hiddenUpdates,s=e[i],s===null?e[i]=[t]:s.push(t),t.lane=n|536870912),l):null}function bo(e){if(50<Ri)throw Ri=0,dl=null,Error(r(185));for(var t=e.return;t!==null;)e=t,t=e.return;return e.tag===3?e.stateNode:null}var cs={};function l0(e,t,n,s){this.tag=e,this.key=n,this.sibling=this.child=this.return=this.stateNode=this.type=this.elementType=null,this.index=0,this.refCleanup=this.ref=null,this.pendingProps=t,this.dependencies=this.memoizedState=this.updateQueue=this.memoizedProps=null,this.mode=s,this.subtreeFlags=this.flags=0,this.deletions=null,this.childLanes=this.lanes=0,this.alternate=null}function St(e,t,n,s){return new l0(e,t,n,s)}function pr(e){return e=e.prototype,!(!e||!e.isReactComponent)}function En(e,t){var n=e.alternate;return n===null?(n=St(e.tag,t,e.key,e.mode),n.elementType=e.elementType,n.type=e.type,n.stateNode=e.stateNode,n.alternate=e,e.alternate=n):(n.pendingProps=t,n.type=e.type,n.flags=0,n.subtreeFlags=0,n.deletions=null),n.flags=e.flags&1206910976,n.childLanes=e.childLanes,n.lanes=e.lanes,n.child=e.child,n.memoizedProps=e.memoizedProps,n.memoizedState=e.memoizedState,n.updateQueue=e.updateQueue,t=e.dependencies,n.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext},n.sibling=e.sibling,n.index=e.index,n.ref=e.ref,n.refCleanup=e.refCleanup,n}function Pu(e,t){e.flags&=1206910978;var n=e.alternate;return n===null?(e.childLanes=0,e.lanes=t,e.child=null,e.subtreeFlags=0,e.memoizedProps=null,e.memoizedState=null,e.updateQueue=null,e.dependencies=null,e.stateNode=null):(e.childLanes=n.childLanes,e.lanes=n.lanes,e.child=n.child,e.subtreeFlags=0,e.deletions=null,e.memoizedProps=n.memoizedProps,e.memoizedState=n.memoizedState,e.updateQueue=n.updateQueue,e.type=n.type,t=n.dependencies,e.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext}),e}function jo(e,t,n,s,i,l){var d=0;if(s=e,typeof s=="function")pr(s)&&(d=1);else if(typeof s=="string")d=q1(e,n,st.current)?26:e==="html"||e==="head"||e==="body"?27:5;else e:switch(s){case Qe:return e=St(31,n,t,i),e.elementType=Qe,e.lanes=l,e;case be:return _a(n.children,i,l,t);case re:d=8,i|=24;break;case J:return e=St(12,n,t,i|2),e.elementType=J,e.lanes=l,e;case ne:return e=St(13,n,t,i),e.elementType=ne,e.lanes=l,e;case le:return e=St(19,n,t,i),e.elementType=le,e.lanes=l,e;case Ze:case j:return e=i|32,e=St(30,n,t,e),e.elementType=j,e.lanes=l,e.stateNode={autoName:null,paired:null,clones:null,ref:null},e;default:if(typeof s=="object"&&s!==null)switch(s.$$typeof){case W:d=10;break e;case ve:d=9;break e;case F:d=11;break e;case je:d=14;break e;case Ae:d=16,s=null;break e}d=29,n=Error(r(130,e===null?"null":typeof e,"")),s=null}return t=St(d,n,t,i),t.elementType=e,t.type=s,t.lanes=l,t}function _a(e,t,n,s){return e=St(7,e,s,t),e.lanes=n,e}function fr(e,t,n){return e=St(6,e,null,t),e.lanes=n,e}function $u(e){var t=St(18,null,null,0);return t.stateNode=e,t}function mr(e,t,n){return t=St(4,e.children!==null?e.children:[],e.key,t),t.lanes=n,t.stateNode={containerInfo:e.containerInfo,pendingChildren:null,implementation:e.implementation},t}var eh=new WeakMap;function Ft(e,t){if(typeof e=="object"&&e!==null){var n=eh.get(e);return n!==void 0?n:(t={value:e,source:t,stack:ja(t)},eh.set(e,t),t)}return{value:e,source:t,stack:ja(t)}}var ds=[],us=0,Ao=null,hi=0,Jt=[],Wt=0,Qn=null,fn=1,mn="";function On(e,t){ds[us++]=hi,ds[us++]=Ao,Ao=e,hi=t}function th(e,t,n){Jt[Wt++]=fn,Jt[Wt++]=mn,Jt[Wt++]=Qn,Qn=e;var s=fn;e=mn;var i=32-Dt(s)-1;s&=~(1<<i),n+=1;var l=32-Dt(t)+i;if(30<l){var d=i-i%5;l=(s&(1<<d)-1).toString(32),s>>=d,i-=d,fn=1<<32-Dt(t)+i|n<<i|s,mn=l+e}else fn=1<<l|n<<i|s,mn=e}function ko(e){e.return!==null&&(On(e,1),th(e,1,0))}function gr(e){for(;e===Ao;)Ao=ds[--us],ds[us]=null,hi=ds[--us],ds[us]=null;for(;e===Qn;)Qn=Jt[--Wt],Jt[Wt]=null,mn=Jt[--Wt],Jt[Wt]=null,fn=Jt[--Wt],Jt[Wt]=null}function nh(e,t){Jt[Wt++]=fn,Jt[Wt++]=mn,Jt[Wt++]=Qn,fn=t.id,mn=t.overflow,Qn=e}var lt=null,Ye=null,Te=!1,Zn=null,Pt=!1,vr=Error(r(519));function Kn(e){var t=Error(r(418,1<arguments.length&&arguments[1]!==void 0&&arguments[1]?"text":"HTML",""));throw pi(Ft(t,e)),vr}function ah(e){var t=e.stateNode,n=e.type,s=e.memoizedProps;switch(t[ht]=e,t[Tt]=s,n){case"dialog":Ne("cancel",t),Ne("close",t);break;case"iframe":case"object":case"embed":Ne("load",t);break;case"video":case"audio":for(n=0;n<qi.length;n++)Ne(qi[n],t);break;case"source":Ne("error",t);break;case"img":case"image":case"link":Ne("error",t),Ne("load",t);break;case"details":Ne("toggle",t);break;case"input":Ne("invalid",t),mu(t,s.value,s.defaultValue,s.checked,s.defaultChecked,s.type,s.name,!0);break;case"select":Ne("invalid",t);break;case"textarea":Ne("invalid",t),vu(t,s.value,s.defaultValue,s.children)}n=s.children,typeof n!="string"&&typeof n!="number"&&typeof n!="bigint"||t.textContent===""+n||s.suppressHydrationWarning===!0||Cf(t.textContent,n)?(s.popover!=null&&(Ne("beforetoggle",t),Ne("toggle",t)),s.onScroll!=null&&Ne("scroll",t),s.onScrollEnd!=null&&Ne("scrollend",t),s.onClick!=null&&(t.onclick=pn),t=!0):t=!1,t||Kn(e,!0)}function Co(e){for(lt=e.return;lt;)switch(lt.tag){case 5:case 31:case 13:Pt=!1;return;case 27:case 3:Pt=!0;return;default:lt=lt.return}}function hs(e){if(e!==lt)return!1;if(!Te)return Co(e),Te=!0,!1;var t=e.tag,n;if((n=t!==3&&t!==27)&&((n=t===5)&&(n=e.type,n=!(n!=="form"&&n!=="button")||Zc(e.type,e.memoizedProps)),n=!n),n&&Ye&&Kn(e),Co(e),t===13){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(r(317));Ye=Xf(e)}else if(t===31){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(r(317));Ye=Xf(e)}else t===27?(t=Ye,da(e.type)?(e=nd,nd=null,Ye=e):Ye=t):Ye=lt?en(e.stateNode.nextSibling):null;return!0}function Ea(){Ye=lt=null,Te=!1}function yr(){var e=Zn;return e!==null&&(Et===null?Et=e:Et.push.apply(Et,e),Zn=null),e}function pi(e){Zn===null?Zn=[e]:Zn.push(e)}var xr=Re(null),Oa=null,zn=null;function Fn(e,t,n){we(xr,t._currentValue),t._currentValue=n}function Rn(e){e._currentValue=xr.current,Ie(xr)}function To(e,t,n){for(;e!==null;){var s=e.alternate;if((e.childLanes&t)!==t?(e.childLanes|=t,s!==null&&(s.childLanes|=t)):s!==null&&(s.childLanes&t)!==t&&(s.childLanes|=t),e===n)break;e=e.return}}function wr(e,t,n,s){var i=e.child;for(i!==null&&(i.return=e);i!==null;){var l=i.dependencies;if(l!==null){var d=i.child;l=l.firstContext;e:for(;l!==null;){var p=l;l=i;for(var k=0;k<t.length;k++)if(p.context===t[k]){l.lanes|=n,p=l.alternate,p!==null&&(p.lanes|=n),To(l.return,n,e),s||(d=null);break e}l=p.next}}else if(i.tag===18){if(d=i.return,d===null)throw Error(r(341));d.lanes|=n,l=d.alternate,l!==null&&(l.lanes|=n),To(d,n,e),d=null}else i.tag===13&&i.memoizedState!==null&&i.memoizedState.dehydrated===null?(i.lanes|=n,d=i.alternate,d!==null&&(d.lanes|=n),To(i.return,n,e),d=i.child,d=d!==null?d.sibling:null):d=i.child;if(d!==null)d.return=i;else for(d=i;d!==null;){if(d===e){d=null;break}if(i=d.sibling,i!==null){i.return=d.return,d=i;break}d=d.return}i=d}}function za(e,t,n,s){e=null;for(var i=t,l=!1;i!==null;){if(!l){if((i.flags&524288)!==0)l=!0;else if((i.flags&262144)!==0)break}if(i.tag===10){var d=i.alternate;if(d===null)throw Error(r(387));if(d=d.memoizedProps,d!==null){var p=i.type;Ut(i.pendingProps.value,d.value)||(e!==null?e.push(p):e=[p])}}else if(i===Ve.current){if(d=i.alternate,d===null)throw Error(r(387));d.memoizedState.memoizedState!==i.memoizedState.memoizedState&&(e!==null?e.push(Ls):e=[Ls])}i=i.return}return e!==null&&wr(t,e,n,s),t.flags|=262144,e!==null}function So(e){for(e=e.firstContext;e!==null;){if(!Ut(e.context._currentValue,e.memoizedValue))return!0;e=e.next}return!1}function Ra(e){Oa=e,zn=null,e=e.dependencies,e!==null&&(e.firstContext=null)}function pt(e){return sh(Oa,e)}function No(e,t){return Oa===null&&Ra(e),sh(e,t)}function sh(e,t){var n=t._currentValue;if(t={context:t,memoizedValue:n,next:null},zn===null){if(e===null)throw Error(r(308));zn=t,e.dependencies={lanes:0,firstContext:t},e.flags|=524288}else zn=zn.next=t;return n}var r0=typeof AbortController<"u"?AbortController:function(){var e=[],t=this.signal={aborted:!1,addEventListener:function(n,s){e.push(s)}};this.abort=function(){t.aborted=!0,e.forEach(function(n){return n()})}},c0=o.unstable_scheduleCallback,d0=o.unstable_NormalPriority,et={$$typeof:W,Consumer:null,Provider:null,_currentValue:null,_currentValue2:null,_threadCount:0};function br(){return{controller:new r0,data:new Map,refCount:0}}function fi(e){e.refCount--,e.refCount===0&&c0(d0,function(){e.controller.abort()})}function ih(e,t){if((e.pendingLanes&4194048)!==0){var n=e.transitionTypes;for(n===null&&(n=e.transitionTypes=[]),e=0;e<t.length;e++){var s=t[e];n.indexOf(s)===-1&&n.push(s)}}}var mi=null;function u0(e){var t=e.transitionTypes;return e.transitionTypes=null,t}var gi=null,jr=0,Ma=0,ps=null;function h0(e,t){if(gi===null){var n=gi=[];jr=0,Ma=Uc(),ps={status:"pending",value:void 0,then:function(s){n.push(s)}}}return jr++,t.then(oh,oh),t}function oh(){if(--jr===0&&(mi=null,gi!==null)){ps!==null&&(ps.status="fulfilled");var e=gi;gi=null,Ma=0,ps=null;for(var t=0;t<e.length;t++)(0,e[t])()}}function p0(e,t){var n=[],s={status:"pending",value:null,reason:null,then:function(i){n.push(i)}};return e.then(function(){s.status="fulfilled",s.value=t;for(var i=0;i<n.length;i++)(0,n[i])(t)},function(i){for(s.status="rejected",s.reason=i,i=0;i<n.length;i++)(0,n[i])(void 0)}),s}var lh=ae.S;ae.S=function(e,t){if($p=Mt(),typeof t=="object"&&t!==null&&typeof t.then=="function"&&h0(e,t),mi!==null)for(var n=Os;n!==null;)ih(n,mi),n=n.next;if(n=e.types,n!==null){for(var s=Os;s!==null;)ih(s,n),s=s.next;if(Ma!==0){s=mi,s===null&&(s=mi=[]);for(var i=0;i<n.length;i++){var l=n[i];s.indexOf(l)===-1&&s.push(l)}}}lh!==null&&lh(e,t)};var qa=Re(null);function Ar(){var e=qa.current;return e!==null?e:Ge.pooledCache}function _o(e,t){t===null?we(qa,qa.current):we(qa,t.pool)}function rh(){var e=Ar();return e===null?null:{parent:et._currentValue,pool:e}}var fs=Error(r(460)),kr=Error(r(474)),Eo=Error(r(542)),Oo={then:function(){}};function ch(e){return e=e.status,e==="fulfilled"||e==="rejected"}function dh(e,t,n){switch(n=e[n],n===void 0?e.push(t):n!==t&&(t.then(pn,pn),t=n),t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,hh(e),e===void 0&&!("reason"in t)?Error(r(600)):e;default:if(typeof t.status=="string")t.then(pn,pn);else{if(e=Ge,e!==null&&100<e.shellSuspendCounter)throw Error(r(482));e=t,e.status="pending",e.then(function(s){if(t.status==="pending"){var i=t;i.status="fulfilled",i.value=s}},function(s){if(t.status==="pending"){var i=t;i.status="rejected",i.reason=s}})}switch(t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,hh(e),e}throw La=t,fs}}function Da(e){try{var t=e._init;return t(e._payload)}catch(n){throw n!==null&&typeof n=="object"&&typeof n.then=="function"?(La=n,fs):n}}var La=null;function uh(){if(La===null)throw Error(r(459));var e=La;return La=null,e}function hh(e){if(e===fs||e===Eo)throw Error(r(483))}var ms=null,vi=0;function zo(e){var t=vi;return vi+=1,ms===null&&(ms=[]),dh(ms,e,t)}function Jn(e,t){t=t.props.ref,e.ref=t!==void 0?t:null}function Ro(e,t){throw t.$$typeof===Z?Error(r(525)):(e=Object.prototype.toString.call(t),Error(r(31,e==="[object Object]"?"object with keys {"+Object.keys(t).join(", ")+"}":e)))}function ph(e){function t(z,T){if(e){var M=z.deletions;M===null?(z.deletions=[T],z.flags|=16):M.push(T)}}function n(z,T){if(!e)return null;for(;T!==null;)t(z,T),T=T.sibling;return null}function s(z){for(var T=new Map;z!==null;)z.key===null?T.set(z.index,z):T.set(z.key,z),z=z.sibling;return T}function i(z,T){return z=En(z,T),z.index=0,z.sibling=null,z}function l(z,T,M){return z.index=M,e?(M=z.alternate,M!==null?(M=M.index,M<T?(z.flags|=2,T):M):(z.flags|=134217730,T)):(z.flags|=1048576,T)}function d(z){return e&&z.alternate===null&&(z.flags|=134217730),z}function p(z,T,M,X){return T===null||T.tag!==6?(T=fr(M,z.mode,X),T.return=z,T):(T=i(T,M),T.return=z,T)}function k(z,T,M,X){var te=M.type;return te===be?(z=L(z,T,M.props.children,X,M.key),Jn(z,M),z):T!==null&&(T.elementType===te||typeof te=="object"&&te!==null&&te.$$typeof===Ae&&Da(te)===T.type)?(T=i(T,M.props),Jn(T,M),T.return=z,T):(T=jo(M.type,M.key,M.props,null,z.mode,X),Jn(T,M),T.return=z,T)}function R(z,T,M,X){return T===null||T.tag!==4||T.stateNode.containerInfo!==M.containerInfo||T.stateNode.implementation!==M.implementation?(T=mr(M,z.mode,X),T.return=z,T):(T=i(T,M.children||[]),T.return=z,T)}function L(z,T,M,X,te){return T===null||T.tag!==7?(T=_a(M,z.mode,X,te),T.return=z,T):(T=i(T,M),T.return=z,T)}function V(z,T,M){if(typeof T=="string"&&T!==""||typeof T=="number"||typeof T=="bigint")return T=fr(""+T,z.mode,M),T.return=z,T;if(typeof T=="object"&&T!==null){switch(T.$$typeof){case $:return M=jo(T.type,T.key,T.props,null,z.mode,M),Jn(M,T),M.return=z,M;case oe:return T=mr(T,z.mode,M),T.return=z,T;case Ae:return T=Da(T),V(z,T,M)}if(xe(T)||ie(T))return T=_a(T,z.mode,M,null),T.return=z,T;if(typeof T.then=="function")return V(z,zo(T),M);if(T.$$typeof===W)return V(z,No(z,T),M);Ro(z,T)}return null}function O(z,T,M,X){var te=T!==null?T.key:null;if(typeof M=="string"&&M!==""||typeof M=="number"||typeof M=="bigint")return te!==null?null:p(z,T,""+M,X);if(typeof M=="object"&&M!==null){switch(M.$$typeof){case $:return M.key===te?k(z,T,M,X):null;case oe:return M.key===te?R(z,T,M,X):null;case Ae:return M=Da(M),O(z,T,M,X)}if(xe(M)||ie(M))return te!==null?null:L(z,T,M,X,null);if(typeof M.then=="function")return O(z,T,zo(M),X);if(M.$$typeof===W)return O(z,T,No(z,M),X);Ro(z,M)}return null}function q(z,T,M,X,te){if(typeof X=="string"&&X!==""||typeof X=="number"||typeof X=="bigint")return z=z.get(M)||null,p(T,z,""+X,te);if(typeof X=="object"&&X!==null){switch(X.$$typeof){case $:return z=z.get(X.key===null?M:X.key)||null,k(T,z,X,te);case oe:return z=z.get(X.key===null?M:X.key)||null,R(T,z,X,te);case Ae:return X=Da(X),q(z,T,M,X,te)}if(xe(X)||ie(X))return z=z.get(M)||null,L(T,z,X,te,null);if(typeof X.then=="function")return q(z,T,M,zo(X),te);if(X.$$typeof===W)return q(z,T,M,No(T,X),te);Ro(T,X)}return null}function P(z,T,M,X){for(var te=null,Ee=null,ue=T,pe=T=0,at=null;ue!==null&&pe<M.length;pe++){ue.index>pe?(at=ue,ue=null):at=ue.sibling;var Oe=O(z,ue,M[pe],X);if(Oe===null){ue===null&&(ue=at);break}e&&ue&&Oe.alternate===null&&t(z,ue),T=l(Oe,T,pe),Ee===null?te=Oe:Ee.sibling=Oe,Ee=Oe,ue=at}if(pe===M.length)return n(z,ue),Te&&On(z,pe),te;if(ue===null){for(;pe<M.length;pe++)ue=V(z,M[pe],X),ue!==null&&(T=l(ue,T,pe),Ee===null?te=ue:Ee.sibling=ue,Ee=ue);return Te&&On(z,pe),te}for(ue=s(ue);pe<M.length;pe++)at=q(ue,z,pe,M[pe],X),at!==null&&(e&&(Oe=at.alternate,Oe!==null&&ue.delete(Oe.key===null?pe:Oe.key)),T=l(at,T,pe),Ee===null?te=at:Ee.sibling=at,Ee=at);return e&&ue.forEach(function(ma){return t(z,ma)}),Te&&On(z,pe),te}function se(z,T,M,X){if(M==null)throw Error(r(151));for(var te=null,Ee=null,ue=T,pe=T=0,at=null,Oe=M.next();ue!==null&&!Oe.done;pe++,Oe=M.next()){ue.index>pe?(at=ue,ue=null):at=ue.sibling;var ma=O(z,ue,Oe.value,X);if(ma===null){ue===null&&(ue=at);break}e&&ue&&ma.alternate===null&&t(z,ue),T=l(ma,T,pe),Ee===null?te=ma:Ee.sibling=ma,Ee=ma,ue=at}if(Oe.done)return n(z,ue),Te&&On(z,pe),te;if(ue===null){for(;!Oe.done;pe++,Oe=M.next())Oe=V(z,Oe.value,X),Oe!==null&&(T=l(Oe,T,pe),Ee===null?te=Oe:Ee.sibling=Oe,Ee=Oe);return Te&&On(z,pe),te}for(ue=s(ue);!Oe.done;pe++,Oe=M.next())Oe=q(ue,z,pe,Oe.value,X),Oe!==null&&(e&&(at=Oe.alternate,at!==null&&ue.delete(at.key===null?pe:at.key)),T=l(Oe,T,pe),Ee===null?te=Oe:Ee.sibling=Oe,Ee=Oe);return e&&ue.forEach(function(Z1){return t(z,Z1)}),Te&&On(z,pe),te}function Ce(z,T,M,X){if(typeof M=="object"&&M!==null&&M.type===be&&M.key===null&&M.props.ref===void 0&&(M=M.props.children),typeof M=="object"&&M!==null){switch(M.$$typeof){case $:e:{for(var te=M.key;T!==null;){if(T.key===te){if(te=M.type,te===be){if(T.tag===7){n(z,T.sibling),X=i(T,M.props.children),Jn(X,M),X.return=z,z=X;break e}}else if(T.elementType===te||typeof te=="object"&&te!==null&&te.$$typeof===Ae&&Da(te)===T.type){n(z,T.sibling),X=i(T,M.props),Jn(X,M),X.return=z,z=X;break e}n(z,T);break}else t(z,T);T=T.sibling}M.type===be?(X=_a(M.props.children,z.mode,X,M.key),Jn(X,M),X.return=z,z=X):(X=jo(M.type,M.key,M.props,null,z.mode,X),Jn(X,M),X.return=z,z=X)}return d(z);case oe:e:{for(te=M.key;T!==null;){if(T.key===te)if(T.tag===4&&T.stateNode.containerInfo===M.containerInfo&&T.stateNode.implementation===M.implementation){n(z,T.sibling),X=i(T,M.children||[]),X.return=z,z=X;break e}else{n(z,T);break}else t(z,T);T=T.sibling}X=mr(M,z.mode,X),X.return=z,z=X}return d(z);case Ae:return M=Da(M),Ce(z,T,M,X)}if(xe(M))return P(z,T,M,X);if(ie(M)){if(te=ie(M),typeof te!="function")throw Error(r(150));return M=te.call(M),se(z,T,M,X)}if(typeof M.then=="function")return Ce(z,T,zo(M),X);if(M.$$typeof===W)return Ce(z,T,No(z,M),X);Ro(z,M)}return typeof M=="string"&&M!==""||typeof M=="number"||typeof M=="bigint"?(M=""+M,T!==null&&T.tag===6?(n(z,T.sibling),X=i(T,M),X.return=z,z=X):(n(z,T),X=fr(M,z.mode,X),X.return=z,z=X),d(z)):n(z,T)}return function(z,T,M,X){try{vi=0;var te=Ce(z,T,M,X);return ms=null,te}catch(ue){if(ue===fs||ue===Eo)throw ue;var Ee=St(29,ue,null,z.mode);return Ee.lanes=X,Ee.return=z,Ee}}}var Ua=ph(!0),fh=ph(!1),Wn=!1;function Cr(e){e.updateQueue={baseState:e.memoizedState,firstBaseUpdate:null,lastBaseUpdate:null,shared:{pending:null,lanes:0,hiddenCallbacks:null},callbacks:null}}function Tr(e,t){e=e.updateQueue,t.updateQueue===e&&(t.updateQueue={baseState:e.baseState,firstBaseUpdate:e.firstBaseUpdate,lastBaseUpdate:e.lastBaseUpdate,shared:e.shared,callbacks:null})}function Pn(e){return{lane:e,tag:0,payload:null,callback:null,next:null}}function $n(e,t,n){var s=e.updateQueue;if(s===null)return null;if(s=s.shared,(Me&2)!==0){var i=s.pending;return i===null?t.next=t:(t.next=i.next,i.next=t),s.pending=t,t=bo(e),Wu(e,null,n),t}return wo(e,s,t,n),bo(e)}function yi(e,t,n){if(t=t.updateQueue,t!==null&&(t=t.shared,(n&4194048)!==0)){var s=t.lanes;s&=e.pendingLanes,n|=s,t.lanes=n,tu(e,n)}}function Sr(e,t){var n=e.updateQueue,s=e.alternate;if(s!==null&&(s=s.updateQueue,n===s)){var i=null,l=null;if(n=n.firstBaseUpdate,n!==null){do{var d={lane:n.lane,tag:n.tag,payload:n.payload,callback:null,next:null};l===null?i=l=d:l=l.next=d,n=n.next}while(n!==null);l===null?i=l=t:l=l.next=t}else i=l=t;n={baseState:s.baseState,firstBaseUpdate:i,lastBaseUpdate:l,shared:s.shared,callbacks:s.callbacks},e.updateQueue=n;return}e=n.lastBaseUpdate,e===null?n.firstBaseUpdate=t:e.next=t,n.lastBaseUpdate=t}var Nr=!1;function xi(){if(Nr){var e=ps;if(e!==null)throw e}}function wi(e,t,n,s){Nr=!1;var i=e.updateQueue;Wn=!1;var l=i.firstBaseUpdate,d=i.lastBaseUpdate,p=i.shared.pending;if(p!==null){i.shared.pending=null;var k=p,R=k.next;k.next=null,d===null?l=R:d.next=R,d=k;var L=e.alternate;L!==null&&(L=L.updateQueue,p=L.lastBaseUpdate,p!==d&&(p===null?L.firstBaseUpdate=R:p.next=R,L.lastBaseUpdate=k))}if(l!==null){var V=i.baseState;d=0,L=R=k=null,p=l;do{var O=p.lane&-536870913,q=O!==p.lane;if(q?(_e&O)===O:(s&O)===O){O!==0&&O===Ma&&(Nr=!0),L!==null&&(L=L.next={lane:0,tag:p.tag,payload:p.payload,callback:null,next:null});e:{var P=e,se=p;O=t;var Ce=n;switch(se.tag){case 1:if(P=se.payload,typeof P=="function"){V=P.call(Ce,V,O);break e}V=P;break e;case 3:P.flags=P.flags&-65537|128;case 0:if(P=se.payload,O=typeof P=="function"?P.call(Ce,V,O):P,O==null)break e;V=U({},V,O);break e;case 2:Wn=!0}}O=p.callback,O!==null&&(e.flags|=64,q&&(e.flags|=8192),q=i.callbacks,q===null?i.callbacks=[O]:q.push(O))}else q={lane:O,tag:p.tag,payload:p.payload,callback:p.callback,next:null},L===null?(R=L=q,k=V):L=L.next=q,d|=O;if(p=p.next,p===null){if(p=i.shared.pending,p===null)break;q=p,p=q.next,q.next=null,i.lastBaseUpdate=q,i.shared.pending=null}}while(!0);L===null&&(k=V),i.baseState=k,i.firstBaseUpdate=R,i.lastBaseUpdate=L,l===null&&(i.shared.lanes=0),oa|=d,e.lanes=d,e.memoizedState=V}}function mh(e,t){if(typeof e!="function")throw Error(r(191,e));e.call(t)}function gh(e,t){var n=e.callbacks;if(n!==null)for(e.callbacks=null,e=0;e<n.length;e++)mh(n[e],t)}var ea=Re(null),Mo=Re(0);function vh(e,t){e=Un,we(Mo,e),we(ea,t),Un=e|t.baseLanes}function _r(){we(Mo,Un),we(ea,ea.current)}function Er(){Un=Mo.current,Ie(ea),Ie(Mo)}var ft=Re(null),xt=null;function ta(e){var t=e.alternate;we(mt,mt.current&1),we(ft,e),xt===null&&(t===null||ea.current!==null||t.memoizedState!==null)&&(xt=e)}function Or(e){we(mt,mt.current),we(ft,e),xt===null&&(xt=e)}function yh(e){e.tag===22?(we(mt,mt.current),we(ft,e),xt===null&&(xt=e)):na()}function na(){we(mt,mt.current),we(ft,ft.current)}function Ht(e){Ie(ft),xt===e&&(xt=null),Ie(mt)}var mt=Re(0);function bi(e,t){we(ft,ft.current),we(mt,t)}function zr(e){Ie(mt),Ie(ft),xt===e&&(xt=null)}function qo(e){for(var t=e;t!==null;){if(t.tag===13){var n=t.memoizedState;if(n!==null&&(n=n.dehydrated,n===null||ed(n)||td(n)))return t}else if(t.tag===19&&t.memoizedProps.revealOrder!=="independent"){if((t.flags&128)!==0)return t}else if(t.child!==null){t.child.return=t,t=t.child;continue}if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return null;t=t.return}t.sibling.return=t.return,t=t.sibling}return null}var Mn=0,ke=null,He=null,tt=null,Do=!1,gs=!1,Ha=!1,Lo=0,ji=0,vs=null,f0=0;function We(){throw Error(r(321))}function Rr(e,t){if(t===null)return!1;for(var n=0;n<t.length&&n<e.length;n++)if(!Ut(e[n],t[n]))return!1;return!0}function Mr(e,t,n,s,i,l){return Mn=l,ke=t,t.memoizedState=null,t.updateQueue=null,t.lanes=0,ae.H=e===null||e.memoizedState===null?tp:np,Ha=!1,l=n(s,i),Ha=!1,gs&&(l=wh(t,n,s,i)),xh(e),l}function xh(e){ae.H=Xo;var t=He!==null&&He.next!==null;if(Mn=0,tt=He=ke=null,Do=!1,ji=0,vs=null,t)throw Error(r(300));e===null||nt||(e=e.dependencies,e!==null&&So(e)&&(nt=!0))}function wh(e,t,n,s){ke=e;var i=0;do{if(gs&&(vs=null),ji=0,gs=!1,25<=i)throw Error(r(301));if(i+=1,tt=He=null,e.updateQueue!=null){var l=e.updateQueue;l.lastEffect=null,l.events=null,l.stores=null,l.memoCache!=null&&(l.memoCache.index=0)}ae.H=j0,l=t(n,s)}while(gs);return l}function m0(){var e=ae.H,t=e.useState()[0];return t=typeof t.then=="function"?Ai(t):t,e=e.useState()[0],(He!==null?He.memoizedState:null)!==e&&(ke.flags|=1024),t}function qr(){var e=Lo!==0;return Lo=0,e}function Dr(e,t,n){t.updateQueue=e.updateQueue,t.flags&=-2053,e.lanes&=~n}function Lr(e){if(Do){for(e=e.memoizedState;e!==null;){var t=e.queue;t!==null&&(t.pending=null),e=e.next}Do=!1}Mn=0,tt=He=ke=null,gs=!1,ji=Lo=0,vs=null}function jt(){var e={memoizedState:null,baseState:null,baseQueue:null,queue:null,next:null};return tt===null?ke.memoizedState=tt=e:tt=tt.next=e,tt}function $e(){if(He===null){var e=ke.alternate;e=e!==null?e.memoizedState:null}else e=He.next;var t=tt===null?ke.memoizedState:tt.next;if(t!==null)tt=t,He=e;else{if(e===null)throw ke.alternate===null?Error(r(467)):Error(r(310));He=e,e={memoizedState:He.memoizedState,baseState:He.baseState,baseQueue:He.baseQueue,queue:He.queue,next:null},tt===null?ke.memoizedState=tt=e:tt=tt.next=e}return tt}function Uo(){return{lastEffect:null,events:null,stores:null,memoCache:null}}function Ai(e){var t=ji;return ji+=1,vs===null&&(vs=[]),e=dh(vs,e,t),t=ke,(tt===null?t.memoizedState:tt.next)===null&&(t=t.alternate,ae.H=t===null||t.memoizedState===null?tp:np),e}function Ho(e){if(e!==null&&typeof e=="object"){if(typeof e.then=="function")return Ai(e);if(e.$$typeof===I)return;if(e.$$typeof===W)return pt(e)}throw Error(r(438,String(e)))}function Ur(e){var t=null,n=ke.updateQueue;if(n!==null&&(t=n.memoCache),t==null){var s=ke.alternate;s!==null&&(s=s.updateQueue,s!==null&&(s=s.memoCache,s!=null&&(t={data:s.data.map(function(i){return i.slice()}),index:0})))}if(t==null&&(t={data:[],index:0}),n===null&&(n=Uo(),ke.updateQueue=n),n.memoCache=t,n=t.data[t.index],n===void 0)for(n=t.data[t.index]=Array(e),s=0;s<e;s++)n[s]=dt;return t.index++,n}function qn(e,t){return typeof t=="function"?t(e):t}function Io(e){var t=$e();return Hr(t,He,e)}function Hr(e,t,n){var s=e.queue;if(s===null)throw Error(r(311));s.lastRenderedReducer=n;var i=e.baseQueue,l=s.pending;if(l!==null){if(i!==null){var d=i.next;i.next=l.next,l.next=d}t.baseQueue=i=l,s.pending=null}if(l=e.baseState,i===null)e.memoizedState=l;else{t=i.next;var p=d=null,k=null,R=t,L=!1;do{var V=R.lane&-536870913;if(V!==R.lane?(_e&V)===V:(Mn&V)===V){var O=R.revertLane;if(O===0)k!==null&&(k=k.next={lane:0,revertLane:0,gesture:null,action:R.action,hasEagerState:R.hasEagerState,eagerState:R.eagerState,next:null}),V===Ma&&(L=!0);else if((Mn&O)===O){R=R.next,O===Ma&&(L=!0);continue}else V={lane:0,revertLane:R.revertLane,gesture:null,action:R.action,hasEagerState:R.hasEagerState,eagerState:R.eagerState,next:null},k===null?(p=k=V,d=l):k=k.next=V,ke.lanes|=O,oa|=O;V=R.action,Ha&&n(l,V),l=R.hasEagerState?R.eagerState:n(l,V)}else O={lane:V,revertLane:R.revertLane,gesture:R.gesture,action:R.action,hasEagerState:R.hasEagerState,eagerState:R.eagerState,next:null},k===null?(p=k=O,d=l):k=k.next=O,ke.lanes|=V,oa|=V;R=R.next}while(R!==null&&R!==t);if(k===null?d=l:k.next=p,!Ut(l,e.memoizedState)&&(nt=!0,L&&(n=ps,n!==null)))throw n;e.memoizedState=l,e.baseState=d,e.baseQueue=k,s.lastRenderedState=l}return i===null&&(s.lanes=0),[e.memoizedState,s.dispatch]}function Ir(e){var t=$e(),n=t.queue;if(n===null)throw Error(r(311));n.lastRenderedReducer=e;var s=n.dispatch,i=n.pending,l=t.memoizedState;if(i!==null){n.pending=null;var d=i=i.next;do l=e(l,d.action),d=d.next;while(d!==i);Ut(l,t.memoizedState)||(nt=!0),t.memoizedState=l,t.baseQueue===null&&(t.baseState=l),n.lastRenderedState=l}return[l,s]}function bh(e,t,n){var s=ke,i=$e(),l=Te;if(l){if(n===void 0)throw Error(r(407));n=n()}else n=t();var d=!Ut((He||i).memoizedState,n);if(d&&(i.memoizedState=n,nt=!0),i=i.queue,Yr(kh.bind(null,s,i,e),[e]),e=i.getSnapshot!==t||d||tt!==null&&(tt.memoizedState.tag&1)!==0,ys(e?9:8,{destroy:void 0},Ah.bind(null,s,i,n,t),null),e){if(s.flags|=2048,Ge===null)throw Error(r(349));l||(Mn&127)!==0||jh(s,t,n)}return n}function jh(e,t,n){e.flags|=16384,e={getSnapshot:t,value:n},t=ke.updateQueue,t===null?(t=Uo(),ke.updateQueue=t,t.stores=[e]):(n=t.stores,n===null?t.stores=[e]:n.push(e))}function Ah(e,t,n,s){t.value=n,t.getSnapshot=s,Ch(t)&&Th(e)}function kh(e,t,n){return n(function(){Ch(t)&&Th(e)})}function Ch(e){var t=e.getSnapshot;e=e.value;try{var n=t();return!Ut(e,n)}catch{return!0}}function Th(e){var t=Na(e,2);t!==null&&Ot(t,e,2)}function Gr(e){var t=jt();if(typeof e=="function"){var n=e;if(e=n(),Ha){Yn(!0);try{n()}finally{Yn(!1)}}}return t.memoizedState=t.baseState=e,t.queue={pending:null,lanes:0,dispatch:null,lastRenderedReducer:qn,lastRenderedState:e},t}function Sh(e,t,n,s){return e.baseState=n,Hr(e,He,typeof s=="function"?s:qn)}function g0(e,t,n,s,i){if(Yo(e))throw Error(r(485));if(e=t.action,e!==null){var l={payload:i,action:e,next:null,isTransition:!0,status:"pending",value:null,reason:null,listeners:[],then:function(d){l.listeners.push(d)}};ae.T!==null?n(!0):l.isTransition=!1,s(l),n=t.pending,n===null?(l.next=t.pending=l,Nh(t,l)):(l.next=n.next,t.pending=n.next=l)}}function Nh(e,t){var n=t.action,s=t.payload,i=e.state;if(t.isTransition){var l=ae.T,d={};d.types=l!==null?l.types:null,ae.T=d;try{var p=n(i,s),k=ae.S;k!==null&&k(d,p),_h(e,t,p)}catch(R){Br(e,t,R)}finally{l!==null&&d.types!==null&&(l.types=d.types),ae.T=l}}else try{l=n(i,s),_h(e,t,l)}catch(R){Br(e,t,R)}}function _h(e,t,n){n!==null&&typeof n=="object"&&typeof n.then=="function"?n.then(function(s){Eh(e,t,s)},function(s){return Br(e,t,s)}):Eh(e,t,n)}function Eh(e,t,n){t.status="fulfilled",t.value=n,Oh(t),e.state=n,t=e.pending,t!==null&&(n=t.next,n===t?e.pending=null:(n=n.next,t.next=n,Nh(e,n)))}function Br(e,t,n){var s=e.pending;if(e.pending=null,s!==null){s=s.next;do t.status="rejected",t.reason=n,Oh(t),t=t.next;while(t!==s)}e.action=null}function Oh(e){e=e.listeners;for(var t=0;t<e.length;t++)(0,e[t])()}function zh(e,t){return t}function Rh(e,t){if(Te){var n=Ge.formState;if(n!==null){e:{var s=ke;if(Te){if(Ye){t:{for(var i=Ye,l=Pt;i.nodeType!==8;){if(!l){i=null;break t}if(i=en(i.nextSibling),i===null){i=null;break t}}l=i.data,i=l==="F!"||l==="F"?i:null}if(i){Ye=en(i.nextSibling),s=i.data==="F!";break e}}Kn(s)}s=!1}s&&(t=n[0])}}return n=jt(),n.memoizedState=n.baseState=t,s={pending:null,lanes:0,dispatch:null,lastRenderedReducer:zh,lastRenderedState:t},n.queue=s,n=Ph.bind(null,ke,s),s.dispatch=n,s=Gr(!1),l=Kr.bind(null,ke,!1,s.queue),s=jt(),i={state:t,dispatch:null,action:e,pending:null},s.queue=i,n=g0.bind(null,ke,i,l,n),i.dispatch=n,s.memoizedState=e,[t,n,!1]}function Mh(e){var t=$e();return qh(t,He,e)}function qh(e,t,n){if(t=Hr(e,t,zh)[0],e=Io(qn)[0],typeof t=="object"&&t!==null&&typeof t.then=="function")try{var s=Ai(t)}catch(d){throw d===fs?Eo:d}else s=t;t=$e();var i=t.queue,l=i.dispatch;return n!==t.memoizedState&&(ke.flags|=2048,ys(9,{destroy:void 0},v0.bind(null,i,n),null)),[s,l,e]}function v0(e,t){e.action=t}function Dh(e){var t=$e(),n=He;if(n!==null)return qh(t,n,e);$e(),t=t.memoizedState,n=$e();var s=n.queue.dispatch;return n.memoizedState=e,[t,s,!1]}function ys(e,t,n,s){return e={tag:e,create:n,deps:s,inst:t,next:null},t=ke.updateQueue,t===null&&(t=Uo(),ke.updateQueue=t),n=t.lastEffect,n===null?t.lastEffect=e.next=e:(s=n.next,n.next=e,e.next=s,t.lastEffect=e),e}function Lh(){return $e().memoizedState}function Go(e,t,n,s){var i=jt();ke.flags|=e,i.memoizedState=ys(1|t,{destroy:void 0},n,s===void 0?null:s)}function Bo(e,t,n,s){var i=$e();s=s===void 0?null:s;var l=i.memoizedState.inst;He!==null&&s!==null&&Rr(s,He.memoizedState.deps)?i.memoizedState=ys(t,l,n,s):(ke.flags|=e,i.memoizedState=ys(1|t,l,n,s))}function Uh(e,t){Go(8390656,8,e,t)}function Yr(e,t){Bo(2048,8,e,t)}function y0(e){ke.flags|=4;var t=ke.updateQueue;if(t===null)t=Uo(),ke.updateQueue=t,t.events=[e];else{var n=t.events;n===null?t.events=[e]:n.push(e)}}function Hh(e){var t=$e().memoizedState;return y0({ref:t,nextImpl:e}),function(){if((Me&2)!==0)throw Error(r(440));return t.impl.apply(void 0,arguments)}}function Ih(e,t){return Bo(4,2,e,t)}function Gh(e,t){return Bo(4,4,e,t)}function Bh(e,t){if(typeof t=="function"){e=e();var n=t(e);return function(){typeof n=="function"?n():t(null)}}if(t!=null)return e=e(),t.current=e,function(){t.current=null}}function Yh(e,t,n){n=n!=null?n.concat([e]):null,Bo(4,4,Bh.bind(null,t,e),n)}function Xr(){}function Xh(e,t){var n=$e();t=t===void 0?null:t;var s=n.memoizedState;return t!==null&&Rr(t,s[1])?s[0]:(n.memoizedState=[e,t],e)}function Vh(e,t){var n=$e();t=t===void 0?null:t;var s=n.memoizedState;if(t!==null&&Rr(t,s[1]))return s[0];if(s=e(),Ha){Yn(!0);try{e()}finally{Yn(!1)}}return n.memoizedState=[s,t],s}function Vr(e,t,n){return n===void 0||(Mn&1073741824)!==0&&(_e&261930)===0?e.memoizedState=t:(e.memoizedState=n,e=tf(),ke.lanes|=e,oa|=e,n)}function Qh(e,t,n,s){return Ut(n,t)?n:ea.current!==null?(e=Vr(e,n,s),Ut(e,t)||(nt=!0),e):(Mn&106)===0||(Mn&1073741824)!==0&&(_e&261930)===0?(nt=!0,e.memoizedState=n):(e=tf(),ke.lanes|=e,oa|=e,t)}function Zh(e,t,n,s,i){var l=he.p;he.p=l!==0&&8>l?l:8;var d=ae.T,p={};p.types=d!==null?d.types:null,ae.T=p,Kr(e,!1,t,n);try{var k=i(),R=ae.S;if(R!==null&&R(p,k),k!==null&&typeof k=="object"&&typeof k.then=="function"){var L=p0(k,s);ki(e,t,L,Yt(e))}else ki(e,t,s,Yt(e))}catch(V){ki(e,t,{then:function(){},status:"rejected",reason:V},Yt())}finally{he.p=l,d!==null&&p.types!==null&&(d.types=p.types),ae.T=d}}function x0(){}function Qr(e,t,n,s){if(e.tag!==5)throw Error(r(476));var i=Kh(e).queue;Zh(e,i,t,ut,n===null?x0:function(){return Fh(e),n(s)})}function Kh(e){var t=e.memoizedState;if(t!==null)return t;t={memoizedState:ut,baseState:ut,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:qn,lastRenderedState:ut},next:null};var n={};return t.next={memoizedState:n,baseState:n,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:qn,lastRenderedState:n},next:null},e.memoizedState=t,e=e.alternate,e!==null&&(e.memoizedState=t),t}function Fh(e){var t=Kh(e);t.next===null&&(t=e.alternate.memoizedState),ki(e,t.next.queue,{},Yt())}function Zr(){return pt(Ls)}function Jh(){return $e().memoizedState}function Wh(){return $e().memoizedState}function w0(e){for(var t=e.return;t!==null;){switch(t.tag){case 24:case 3:var n=Yt();e=Pn(n);var s=$n(t,e,n);s!==null&&(Ot(s,t,n),yi(s,t,n)),t={cache:br()},e.payload=t;return}t=t.return}}function b0(e,t,n){var s=Yt();n={lane:s,revertLane:0,gesture:null,action:n,hasEagerState:!1,eagerState:null,next:null},Yo(e)?$h(t,n):(n=hr(e,t,n,s),n!==null&&(Ot(n,e,s),ep(n,t,s)))}function Ph(e,t,n){var s=Yt();ki(e,t,n,s)}function ki(e,t,n,s){var i={lane:s,revertLane:0,gesture:null,action:n,hasEagerState:!1,eagerState:null,next:null};if(Yo(e))$h(t,i);else{var l=e.alternate;if(e.lanes===0&&(l===null||l.lanes===0)&&(l=t.lastRenderedReducer,l!==null))try{var d=t.lastRenderedState,p=l(d,n);if(i.hasEagerState=!0,i.eagerState=p,Ut(p,d))return wo(e,t,i,0),Ge===null&&xo(),!1}catch{}if(n=hr(e,t,i,s),n!==null)return Ot(n,e,s),ep(n,t,s),!0}return!1}function Kr(e,t,n,s){if(s={lane:2,revertLane:Uc(),gesture:null,action:s,hasEagerState:!1,eagerState:null,next:null},Yo(e)){if(t)throw Error(r(479))}else t=hr(e,n,s,2),t!==null&&Ot(t,e,2)}function Yo(e){var t=e.alternate;return e===ke||t!==null&&t===ke}function $h(e,t){gs=Do=!0;var n=e.pending;n===null?t.next=t:(t.next=n.next,n.next=t),e.pending=t}function ep(e,t,n){if((n&4194048)!==0){var s=t.lanes;s&=e.pendingLanes,n|=s,t.lanes=n,tu(e,n)}}var Xo={readContext:pt,use:Ho,useCallback:We,useContext:We,useEffect:We,useImperativeHandle:We,useLayoutEffect:We,useInsertionEffect:We,useMemo:We,useReducer:We,useRef:We,useState:We,useDebugValue:We,useDeferredValue:We,useTransition:We,useSyncExternalStore:We,useId:We,useHostTransitionStatus:We,useFormState:We,useActionState:We,useOptimistic:We,useMemoCache:We,useCacheRefresh:We,useEffectEvent:We},tp={readContext:pt,use:Ho,useCallback:function(e,t){return jt().memoizedState=[e,t===void 0?null:t],e},useContext:pt,useEffect:Uh,useImperativeHandle:function(e,t,n){n=n!=null?n.concat([e]):null,Go(4194308,4,Bh.bind(null,t,e),n)},useLayoutEffect:function(e,t){return Go(4194308,4,e,t)},useInsertionEffect:function(e,t){Go(4,2,e,t)},useMemo:function(e,t){var n=jt();t=t===void 0?null:t;var s=e();if(Ha){Yn(!0);try{e()}finally{Yn(!1)}}return n.memoizedState=[s,t],s},useReducer:function(e,t,n){var s=jt();if(n!==void 0){var i=n(t);if(Ha){Yn(!0);try{n(t)}finally{Yn(!1)}}}else i=t;return s.memoizedState=s.baseState=i,e={pending:null,lanes:0,dispatch:null,lastRenderedReducer:e,lastRenderedState:i},s.queue=e,e=e.dispatch=b0.bind(null,ke,e),[s.memoizedState,e]},useRef:function(e){var t=jt();return e={current:e},t.memoizedState=e},useState:function(e){e=Gr(e);var t=e.queue,n=Ph.bind(null,ke,t);return t.dispatch=n,[e.memoizedState,n]},useDebugValue:Xr,useDeferredValue:function(e,t){var n=jt();return Vr(n,e,t)},useTransition:function(){var e=Gr(!1);return e=Zh.bind(null,ke,e.queue,!0,!1),jt().memoizedState=e,[!1,e]},useSyncExternalStore:function(e,t,n){var s=ke,i=jt();if(Te){if(n===void 0)throw Error(r(407));n=n()}else{if(n=t(),Ge===null)throw Error(r(349));(_e&127)!==0||jh(s,t,n)}i.memoizedState=n;var l={value:n,getSnapshot:t};return i.queue=l,Uh(kh.bind(null,s,l,e),[e]),s.flags|=2048,ys(9,{destroy:void 0},Ah.bind(null,s,l,n,t),null),n},useId:function(){var e=jt(),t=Ge.identifierPrefix;if(Te){var n=mn,s=fn;n=(s&~(1<<32-Dt(s)-1)).toString(32)+n,t="_"+t+"R_"+n,n=Lo++,0<n&&(t+="H"+n.toString(32)),t+="_"}else n=f0++,t="_"+t+"r_"+n.toString(32)+"_";return e.memoizedState=t},useHostTransitionStatus:Zr,useFormState:Rh,useActionState:Rh,useOptimistic:function(e){var t=jt();t.memoizedState=t.baseState=e;var n={pending:null,lanes:0,dispatch:null,lastRenderedReducer:null,lastRenderedState:null};return t.queue=n,t=Kr.bind(null,ke,!0,n),n.dispatch=t,[e,t]},useMemoCache:Ur,useCacheRefresh:function(){return jt().memoizedState=w0.bind(null,ke)},useEffectEvent:function(e){var t=jt(),n={impl:e};return t.memoizedState=n,function(){if((Me&2)!==0)throw Error(r(440));return n.impl.apply(void 0,arguments)}}},np={readContext:pt,use:Ho,useCallback:Xh,useContext:pt,useEffect:Yr,useImperativeHandle:Yh,useInsertionEffect:Ih,useLayoutEffect:Gh,useMemo:Vh,useReducer:Io,useRef:Lh,useState:function(){return Io(qn)},useDebugValue:Xr,useDeferredValue:function(e,t){var n=$e();return Qh(n,He.memoizedState,e,t)},useTransition:function(){var e=Io(qn)[0],t=$e().memoizedState;return[typeof e=="boolean"?e:Ai(e),t]},useSyncExternalStore:bh,useId:Jh,useHostTransitionStatus:Zr,useFormState:Mh,useActionState:Mh,useOptimistic:function(e,t){var n=$e();return Sh(n,He,e,t)},useMemoCache:Ur,useCacheRefresh:Wh,useEffectEvent:Hh},j0={readContext:pt,use:Ho,useCallback:Xh,useContext:pt,useEffect:Yr,useImperativeHandle:Yh,useInsertionEffect:Ih,useLayoutEffect:Gh,useMemo:Vh,useReducer:Ir,useRef:Lh,useState:function(){return Ir(qn)},useDebugValue:Xr,useDeferredValue:function(e,t){var n=$e();return He===null?Vr(n,e,t):Qh(n,He.memoizedState,e,t)},useTransition:function(){var e=Ir(qn)[0],t=$e().memoizedState;return[typeof e=="boolean"?e:Ai(e),t]},useSyncExternalStore:bh,useId:Jh,useHostTransitionStatus:Zr,useFormState:Dh,useActionState:Dh,useOptimistic:function(e,t){var n=$e();return He!==null?Sh(n,He,e,t):(n.baseState=e,[e,n.queue.dispatch])},useMemoCache:Ur,useCacheRefresh:Wh,useEffectEvent:Hh};function Fr(e,t,n,s){t=e.memoizedState,n=n(s,t),n=n==null?t:U({},t,n),e.memoizedState=n,e.lanes===0&&(e.updateQueue.baseState=n)}var Jr={enqueueSetState:function(e,t,n){e=e._reactInternals;var s=Yt(),i=Pn(s);i.payload=t,n!=null&&(i.callback=n),t=$n(e,i,s),t!==null&&(Ot(t,e,s),yi(t,e,s))},enqueueReplaceState:function(e,t,n){e=e._reactInternals;var s=Yt(),i=Pn(s);i.tag=1,i.payload=t,n!=null&&(i.callback=n),t=$n(e,i,s),t!==null&&(Ot(t,e,s),yi(t,e,s))},enqueueForceUpdate:function(e,t){e=e._reactInternals;var n=Yt(),s=Pn(n);s.tag=2,t!=null&&(s.callback=t),t=$n(e,s,n),t!==null&&(Ot(t,e,n),yi(t,e,n))}};function ap(e,t,n,s,i,l,d){return e=e.stateNode,typeof e.shouldComponentUpdate=="function"?e.shouldComponentUpdate(s,l,d):t.prototype&&t.prototype.isPureReactComponent?!di(n,s)||!di(i,l):!0}function sp(e,t,n,s){e=t.state,typeof t.componentWillReceiveProps=="function"&&t.componentWillReceiveProps(n,s),typeof t.UNSAFE_componentWillReceiveProps=="function"&&t.UNSAFE_componentWillReceiveProps(n,s),t.state!==e&&Jr.enqueueReplaceState(t,t.state,null)}function Ia(e,t){var n=t;if("ref"in t){n={};for(var s in t)s!=="ref"&&(n[s]=t[s])}if(e=e.defaultProps){n===t&&(n=U({},n));for(var i in e)n[i]===void 0&&(n[i]=e[i])}return n}function ip(e){yo(e)}function op(e){console.error(e)}function lp(e){yo(e)}function Vo(e,t){try{var n=e.onUncaughtError;n(t.value,{componentStack:t.stack})}catch(s){setTimeout(function(){throw s})}}function rp(e,t,n){try{var s=e.onCaughtError;s(n.value,{componentStack:n.stack,errorBoundary:t.tag===1?t.stateNode:null})}catch(i){setTimeout(function(){throw i})}}function Wr(e,t,n){return n=Pn(n),n.tag=3,n.payload={element:null},n.callback=function(){Vo(e,t)},n}function cp(e){return e=Pn(e),e.tag=3,e}function dp(e,t,n,s){var i=n.type.getDerivedStateFromError;if(typeof i=="function"){var l=s.value;e.payload=function(){return i(l)},e.callback=function(){rp(t,n,s)}}var d=n.stateNode;d!==null&&typeof d.componentDidCatch=="function"&&(e.callback=function(){rp(t,n,s),typeof i!="function"&&(la===null?la=new Set([this]):la.add(this));var p=s.stack;this.componentDidCatch(s.value,{componentStack:p!==null?p:""})})}function A0(e,t,n,s,i){if(n.flags|=32768,s!==null&&typeof s=="object"&&typeof s.then=="function"){if(t=n.alternate,t!==null&&za(t,n,i,!0),n=ft.current,n!==null){switch(n.tag){case 31:case 13:case 19:return xt===null?hl():n.alternate===null&&Pe===0&&(Pe=3),n.flags&=-257,n.flags|=65536,n.lanes=i,s===Oo?n.flags|=16384:(t=n.updateQueue,t===null?n.updateQueue=new Set([s]):t.add(s),qc(e,s,i)),!1;case 22:return n.flags|=65536,s===Oo?n.flags|=16384:(t=n.updateQueue,t===null?(t={transitions:null,markerInstances:null,retryQueue:new Set([s])},n.updateQueue=t):(n=t.retryQueue,n===null?t.retryQueue=new Set([s]):n.add(s)),qc(e,s,i)),!1}throw Error(r(435,n.tag))}return qc(e,s,i),hl(),!1}if(Te)return t=ft.current,t!==null?((t.flags&65536)===0&&(t.flags|=256),t.flags|=65536,t.lanes=i,s!==vr&&(e=Error(r(422),{cause:s}),pi(Ft(e,n)))):(s!==vr&&(t=Error(r(423),{cause:s}),pi(Ft(t,n))),e=e.current.alternate,e.flags|=65536,i&=-i,e.lanes|=i,s=Ft(s,n),i=Wr(e.stateNode,s,i),Sr(e,i),Pe!==4&&(Pe=2)),!1;var l=Error(r(520),{cause:s});if(l=Ft(l,n),zi===null?zi=[l]:zi.push(l),Pe!==4&&(Pe=2),t===null)return!0;s=Ft(s,n),n=t;do{switch(n.tag){case 3:return n.flags|=65536,e=i&-i,n.lanes|=e,e=Wr(n.stateNode,s,e),Sr(n,e),!1;case 1:if(t=n.type,l=n.stateNode,(n.flags&128)===0&&(typeof t.getDerivedStateFromError=="function"||l!==null&&typeof l.componentDidCatch=="function"&&(la===null||!la.has(l))))return n.flags|=65536,i&=-i,n.lanes|=i,i=cp(i),dp(i,e,n,s),Sr(n,i),!1;break;case 22:if(n.memoizedState!==null)return n.flags|=65536,!1}n=n.return}while(n!==null);return!1}var Pr=Error(r(461)),nt=!1;function it(e,t,n,s){t.child=e===null?fh(t,null,n,s):Ua(t,e.child,n,s)}function up(e,t,n,s,i){n=n.render;var l=t.ref;if("ref"in s){var d={};for(var p in s)p!=="ref"&&(d[p]=s[p])}else d=s;return Ra(t),s=Mr(e,t,n,d,l,i),p=qr(),e!==null&&!nt?(Dr(e,t,i),Dn(e,t,i)):(Te&&p&&ko(t),t.flags|=1,it(e,t,s,i),t.child)}function hp(e,t,n,s,i){if(e===null){var l=n.type;return typeof l=="function"&&!pr(l)&&l.defaultProps===void 0&&n.compare===null?(t.tag=15,t.type=l,pp(e,t,l,s,i)):(e=jo(n.type,null,s,t,t.mode,i),e.ref=t.ref,e.return=t,t.child=e)}if(l=e.child,!oc(e,i)){var d=l.memoizedProps;if(n=n.compare,n=n!==null?n:di,n(d,s)&&e.ref===t.ref)return Dn(e,t,i)}return t.flags|=1,e=En(l,s),e.ref=t.ref,e.return=t,t.child=e}function pp(e,t,n,s,i){if(e!==null){var l=e.memoizedProps;if(di(l,s)&&e.ref===t.ref)if(nt=!1,t.pendingProps=s=l,oc(e,i))(e.flags&131072)!==0&&(nt=!0);else return t.lanes=e.lanes,Dn(e,t,i)}return $r(e,t,n,s,i)}function fp(e,t,n,s){var i=s.children,l=e!==null?e.memoizedState:null;if(e===null&&t.stateNode===null&&(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),s.mode==="hidden"){if((t.flags&128)!==0){if(l=l!==null?l.baseLanes|n:n,e!==null){for(s=t.child=e.child,i=0;s!==null;)i=i|s.lanes|s.childLanes,s=s.sibling;s=i&~l}else s=0,t.child=null;return mp(e,t,l,n,s)}if((n&536870912)!==0)t.memoizedState={baseLanes:0,cachePool:null},e!==null&&_o(t,l!==null?l.cachePool:null),l!==null?vh(t,l):_r(),yh(t);else return s=t.lanes=536870912,mp(e,t,l!==null?l.baseLanes|n:n,n,s)}else l!==null?(_o(t,l.cachePool),vh(t,l),na(),t.memoizedState=null):(e!==null&&_o(t,null),_r(),na());return it(e,t,i,n),t.child}function Ci(e,t){return e!==null&&e.tag===22||t.stateNode!==null||(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),t.sibling}function mp(e,t,n,s,i){var l=Ar();return l=l===null?null:{parent:et._currentValue,pool:l},t.memoizedState={baseLanes:n,cachePool:l},e!==null&&_o(t,null),_r(),yh(t),e!==null&&za(e,t,s,!0),t.childLanes=i,null}function Qo(e,t){return t=Zo({mode:t.mode,children:t.children},e.mode),t.ref=e.ref,e.child=t,t.return=e,t}function gp(e,t,n){return Ua(t,e.child,null,n),e=Qo(t,t.pendingProps),e.flags|=2,Ht(t),t.memoizedState=null,e}function k0(e,t,n){var s=t.pendingProps,i=(t.flags&128)!==0;if(t.flags&=-129,e===null){if(Te){if(s.mode==="hidden")return e=Qo(t,s),t.lanes=536870912,e.memoizedState={baseLanes:0,cachePool:null},Ci(null,e);if(Or(t),(e=Ye)?(e=Yf(e,Pt),e=e!==null&&e.data==="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:Qn!==null?{id:fn,overflow:mn}:null,retryLane:536870912,hydrationErrors:null},n=$u(e),n.return=t,t.child=n,lt=t,Ye=null)):e=null,e===null)throw Kn(t);return t.lanes=536870912,null}return Qo(t,s)}var l=e.memoizedState;if(l!==null){var d=l.dehydrated;if(Or(t),i)if(t.flags&256)t.flags&=-257,t=gp(e,t,n);else if(t.memoizedState!==null)t.child=e.child,t.flags|=128,t=null;else throw Error(r(558));else if(nt||za(e,t,n,!1),i=(n&e.childLanes)!==0,nt||i){if(ea.current===null){if(s=Ge,s!==null&&(d=nu(s,n),d!==0&&d!==l.retryLane))throw l.retryLane=d,Na(e,d),Ot(s,e,d),Pr;hl()}t=gp(e,t,n)}else e=l.treeContext,Ye=en(d.nextSibling),lt=t,Te=!0,Zn=null,Pt=!1,e!==null&&nh(t,e),t=Qo(t,s),t.flags|=134221824;return t}return e=En(e.child,{mode:s.mode,children:s.children}),e.ref=t.ref,t.child=e,e.return=t,e}function xs(e,t){var n=t.ref;if(n===null)e!==null&&e.ref!==null&&(t.flags|=4194816);else{if(typeof n!="function"&&typeof n!="object")throw Error(r(284));(e===null||e.ref!==n)&&(t.flags|=4194816)}}function $r(e,t,n,s,i){return Ra(t),n=Mr(e,t,n,s,void 0,i),s=qr(),e!==null&&!nt?(Dr(e,t,i),Dn(e,t,i)):(Te&&s&&ko(t),t.flags|=1,it(e,t,n,i),t.child)}function vp(e,t,n,s,i,l){return Ra(t),t.updateQueue=null,n=wh(t,s,n,i),xh(e),s=qr(),e!==null&&!nt?(Dr(e,t,l),Dn(e,t,l)):(Te&&s&&ko(t),t.flags|=1,it(e,t,n,l),t.child)}function yp(e,t,n,s,i){if(Ra(t),t.stateNode===null){var l=cs,d=n.contextType;typeof d=="object"&&d!==null&&(l=pt(d)),l=new n(s,l),t.memoizedState=l.state!==null&&l.state!==void 0?l.state:null,l.updater=Jr,t.stateNode=l,l._reactInternals=t,l=t.stateNode,l.props=s,l.state=t.memoizedState,l.refs={},Cr(t),d=n.contextType,l.context=typeof d=="object"&&d!==null?pt(d):cs,l.state=t.memoizedState,d=n.getDerivedStateFromProps,typeof d=="function"&&(Fr(t,n,d,s),l.state=t.memoizedState),typeof n.getDerivedStateFromProps=="function"||typeof l.getSnapshotBeforeUpdate=="function"||typeof l.UNSAFE_componentWillMount!="function"&&typeof l.componentWillMount!="function"||(d=l.state,typeof l.componentWillMount=="function"&&l.componentWillMount(),typeof l.UNSAFE_componentWillMount=="function"&&l.UNSAFE_componentWillMount(),d!==l.state&&Jr.enqueueReplaceState(l,l.state,null),wi(t,s,l,i),xi(),l.state=t.memoizedState),typeof l.componentDidMount=="function"&&(t.flags|=4194308),s=!0}else if(e===null){l=t.stateNode;var p=t.memoizedProps,k=Ia(n,p);l.props=k;var R=l.context,L=n.contextType;d=cs,typeof L=="object"&&L!==null&&(d=pt(L));var V=n.getDerivedStateFromProps;L=typeof V=="function"||typeof l.getSnapshotBeforeUpdate=="function",p=t.pendingProps!==p,L||typeof l.UNSAFE_componentWillReceiveProps!="function"&&typeof l.componentWillReceiveProps!="function"||(p||R!==d)&&sp(t,l,s,d),Wn=!1;var O=t.memoizedState;l.state=O,wi(t,s,l,i),xi(),R=t.memoizedState,p||O!==R||Wn?(typeof V=="function"&&(Fr(t,n,V,s),R=t.memoizedState),(k=Wn||ap(t,n,k,s,O,R,d))?(L||typeof l.UNSAFE_componentWillMount!="function"&&typeof l.componentWillMount!="function"||(typeof l.componentWillMount=="function"&&l.componentWillMount(),typeof l.UNSAFE_componentWillMount=="function"&&l.UNSAFE_componentWillMount()),typeof l.componentDidMount=="function"&&(t.flags|=4194308)):(typeof l.componentDidMount=="function"&&(t.flags|=4194308),t.memoizedProps=s,t.memoizedState=R),l.props=s,l.state=R,l.context=d,s=k):(typeof l.componentDidMount=="function"&&(t.flags|=4194308),s=!1)}else{l=t.stateNode,Tr(e,t),d=t.memoizedProps,L=Ia(n,d),l.props=L,V=t.pendingProps,O=l.context,R=n.contextType,k=cs,typeof R=="object"&&R!==null&&(k=pt(R)),p=n.getDerivedStateFromProps,(R=typeof p=="function"||typeof l.getSnapshotBeforeUpdate=="function")||typeof l.UNSAFE_componentWillReceiveProps!="function"&&typeof l.componentWillReceiveProps!="function"||(d!==V||O!==k)&&sp(t,l,s,k),Wn=!1,O=t.memoizedState,l.state=O,wi(t,s,l,i),xi();var q=t.memoizedState;d!==V||O!==q||Wn||e!==null&&e.dependencies!==null&&So(e.dependencies)?(typeof p=="function"&&(Fr(t,n,p,s),q=t.memoizedState),(L=Wn||ap(t,n,L,s,O,q,k)||e!==null&&e.dependencies!==null&&So(e.dependencies))?(R||typeof l.UNSAFE_componentWillUpdate!="function"&&typeof l.componentWillUpdate!="function"||(typeof l.componentWillUpdate=="function"&&l.componentWillUpdate(s,q,k),typeof l.UNSAFE_componentWillUpdate=="function"&&l.UNSAFE_componentWillUpdate(s,q,k)),typeof l.componentDidUpdate=="function"&&(t.flags|=4),typeof l.getSnapshotBeforeUpdate=="function"&&(t.flags|=1024)):(typeof l.componentDidUpdate!="function"||d===e.memoizedProps&&O===e.memoizedState||(t.flags|=4),typeof l.getSnapshotBeforeUpdate!="function"||d===e.memoizedProps&&O===e.memoizedState||(t.flags|=1024),t.memoizedProps=s,t.memoizedState=q),l.props=s,l.state=q,l.context=k,s=L):(typeof l.componentDidUpdate!="function"||d===e.memoizedProps&&O===e.memoizedState||(t.flags|=4),typeof l.getSnapshotBeforeUpdate!="function"||d===e.memoizedProps&&O===e.memoizedState||(t.flags|=1024),s=!1)}return l=s,xs(e,t),s=(t.flags&128)!==0,l||s?(l=t.stateNode,n=s&&typeof n.getDerivedStateFromError!="function"?null:l.render(),t.flags|=1,e!==null&&s?(t.child=Ua(t,e.child,null,i),t.child=Ua(t,null,n,i)):it(e,t,n,i),t.memoizedState=l.state,e=t.child):e=Dn(e,t,i),e}function xp(e,t,n,s){return Ea(),t.flags|=256,it(e,t,n,s),t.child}var ec={dehydrated:null,treeContext:null,retryLane:0,hydrationErrors:null};function tc(e){return{baseLanes:e,cachePool:rh()}}function nc(e,t,n){return e=e!==null?e.childLanes&~n:0,t&&(e|=Bt),e}function wp(e,t,n){var s=t.pendingProps,i=!1,l=(t.flags&128)!==0,d;if((d=l)||(d=e!==null&&e.memoizedState===null?!1:(mt.current&2)!==0),d&&(i=!0,t.flags&=-129),d=(t.flags&32)!==0,t.flags&=-33,e===null){if(Te){if(i?ta(t):na(),(e=Ye)?(e=Yf(e,Pt),e=e!==null&&e.data!=="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:Qn!==null?{id:fn,overflow:mn}:null,retryLane:536870912,hydrationErrors:null},n=$u(e),n.return=t,t.child=n,lt=t,Ye=null)):e=null,e===null)throw Kn(t);return td(e)?t.lanes=32:t.lanes=536870912,null}return l=s.children,s=s.fallback,i?(na(),i=t.mode,l=Zo({mode:"hidden",children:l},i),s=_a(s,i,n,null),l.return=t,s.return=t,l.sibling=s,t.child=l,s=t.child,s.memoizedState=tc(n),s.childLanes=nc(e,d,n),t.memoizedState=ec,Ci(null,s)):(ta(t),ac(t,l))}var p=e.memoizedState;if(p!==null){var k=p.dehydrated;if(k!==null)return C0(e,t,l,d,s,k,p,n)}return i?(na(),i=s.fallback,l=t.mode,p=e.child,k=p.sibling,s=En(p,{mode:"hidden",children:s.children}),s.subtreeFlags=p.subtreeFlags&1206910976,k!==null?i=En(k,i):(i=_a(i,l,n,null),i.flags|=2),i.return=t,s.return=t,s.sibling=i,t.child=s,Ci(null,s),s=t.child,i=e.child.memoizedState,i===null?i=tc(n):(l=i.cachePool,l!==null?(p=et._currentValue,l=l.parent!==p?{parent:p,pool:p}:l):l=rh(),i={baseLanes:i.baseLanes|n,cachePool:l}),s.memoizedState=i,s.childLanes=nc(e,d,n),t.memoizedState=ec,Ci(e.child,s)):(ta(t),n=e.child,e=n.sibling,n=En(n,{mode:"visible",children:s.children}),n.return=t,n.sibling=null,e!==null&&(d=t.deletions,d===null?(t.deletions=[e],t.flags|=16):d.push(e)),t.child=n,t.memoizedState=null,n)}function ac(e,t){return t=Zo({mode:"visible",children:t},e.mode),t.return=e,e.child=t}function Zo(e,t){return e=St(22,e,null,t),e.lanes=0,e}function Ko(e,t,n){return Ua(t,e.child,null,n),e=ac(t,t.pendingProps.children),e.flags|=2,t.memoizedState=null,e}function C0(e,t,n,s,i,l,d,p){if(n)return t.flags&256?(ta(t),t.flags&=-257,Ko(e,t,p)):t.memoizedState!==null?(na(),t.child=e.child,t.flags|=128,null):(na(),l=i.fallback,d=t.mode,i=Zo({mode:"visible",children:i.children},d),l=_a(l,d,p,null),l.flags|=2,i.return=t,l.return=t,i.sibling=l,t.child=i,Ua(t,e.child,null,p),i=t.child,i.memoizedState=tc(p),i.childLanes=nc(e,s,p),t.memoizedState=ec,Ci(null,i));if(ta(t),td(l)){if(s=l.nextSibling&&l.nextSibling.dataset,s)var k=s.dgst;return s=k,s!==""&&(i=Error(r(419)),i.stack="",i.digest=s,pi({value:i,source:null,stack:null})),Ko(e,t,p)}if(nt||za(e,t,p,!1),s=(p&e.childLanes)!==0,nt||s){if(ea.current!==null)return Ko(e,t,p);if(s=Ge,s!==null&&(i=nu(s,p),i!==0&&i!==d.retryLane))throw d.retryLane=i,Na(e,i),Ot(s,e,i),Pr;return ed(l)||hl(),Ko(e,t,p)}return ed(l)?(t.flags|=192,t.child=e.child,null):(e=d.treeContext,Ye=en(l.nextSibling),lt=t,Te=!0,Zn=null,Pt=!1,e!==null&&nh(t,e),t=ac(t,i.children),t.flags|=134221824,t)}function bp(e,t,n){e.lanes|=t;var s=e.alternate;s!==null&&(s.lanes|=t),To(e.return,t,n)}function jp(e){for(var t=null;e!==null;){var n=e.alternate;n!==null&&qo(n)===null&&(t=e),e=e.sibling}return t}function Fo(e,t,n,s,i,l){var d=e.memoizedState;d===null?e.memoizedState={isBackwards:t,rendering:null,renderingStartTime:0,last:s,tail:n,tailMode:i,treeForkCount:l}:(d.isBackwards=t,d.rendering=null,d.renderingStartTime=0,d.last=s,d.tail=n,d.tailMode=i,d.treeForkCount=l)}function sc(e){var t=e.child;for(e.child=null;t!==null;){var n=t.sibling;t.sibling=e.child,e.child=t,t=n}}function ic(e,t,n){var s=t.pendingProps,i=s.revealOrder,l=s.tail;s=s.children;var d=mt.current;if(t.flags&128)return bi(t,d),null;var p=(d&2)!==0;if(p?(d=d&1|2,t.flags|=128):d&=1,bi(t,d),i==="backwards"&&e!==null?(sc(e),it(e,t,s,n),sc(e)):it(e,t,s,n),s=Te?hi:0,!p&&e!==null&&(e.flags&128)!==0)e:for(e=t.child;e!==null;){if(e.tag===13)e.memoizedState!==null&&bp(e,n,t);else if(e.tag===19)bp(e,n,t);else if(e.child!==null){e.child.return=e,e=e.child;continue}if(e===t)break e;for(;e.sibling===null;){if(e.return===null||e.return===t)break e;e=e.return}e.sibling.return=e.return,e=e.sibling}switch(i){case"backwards":n=jp(t.child),n===null?(i=t.child,t.child=null):(i=n.sibling,n.sibling=null,sc(t)),Fo(t,!0,i,null,l,s);break;case"unstable_legacy-backwards":for(n=null,i=t.child,t.child=null;i!==null;){if(e=i.alternate,e!==null&&qo(e)===null){t.child=i;break}e=i.sibling,i.sibling=n,n=i,i=e}Fo(t,!0,n,null,l,s);break;case"together":Fo(t,!1,null,null,void 0,s);break;case"independent":t.memoizedState=null;break;default:n=jp(t.child),n===null?(i=t.child,t.child=null):(i=n.sibling,n.sibling=null),Fo(t,!1,i,n,l,s)}return t.child}function Ap(e,t,n){var s=t.pendingProps;return Fn(t,t.type,s.value),it(e,t,s.children,n),t.child}function Dn(e,t,n){if(e!==null&&(t.dependencies=e.dependencies),oa|=t.lanes,(n&t.childLanes)===0)if(e!==null){if(za(e,t,n,!1),(n&t.childLanes)===0)return null}else return null;if(e!==null&&t.child!==e.child)throw Error(r(153));if(t.child!==null){for(e=t.child,n=En(e,e.pendingProps),t.child=n,n.return=t;e.sibling!==null;)e=e.sibling,n=n.sibling=En(e,e.pendingProps),n.return=t;n.sibling=null}return t.child}function oc(e,t){return(e.lanes&t)!==0?!0:(e=e.dependencies,!!(e!==null&&So(e)))}function T0(e,t,n){switch(t.tag){case 3:nn(t,t.stateNode.containerInfo),Fn(t,et,e.memoizedState.cache),Ea();break;case 27:case 5:Js(t);break;case 4:nn(t,t.stateNode.containerInfo);break;case 10:Fn(t,t.type,t.memoizedProps.value);break;case 31:if(t.memoizedState!==null)return t.flags|=128,Or(t),null;break;case 13:var s=t.memoizedState;if(s!==null){if(s.dehydrated!==null)return ta(t),t.flags|=128,null;s=za(e,t,n,!1);var i=t.child.childLanes;return s||(n&i)!==0?wp(e,t,n):(ta(t),e=Dn(e,t,n),e!==null?e.sibling:null)}ta(t);break;case 19:if(t.flags&128)return ic(e,t,n);if(i=(e.flags&128)!==0,s=(n&t.childLanes)!==0,s||(za(e,t,n,!1),s=(n&t.childLanes)!==0),i){if(s)return ic(e,t,n);t.flags|=128}if(i=t.memoizedState,i!==null&&(i.rendering=null,i.tail=null,i.lastEffect=null),bi(t,mt.current),s)break;return null;case 22:return t.lanes=0,fp(e,t,n,t.pendingProps);case 24:Fn(t,et,e.memoizedState.cache)}return Dn(e,t,n)}function kp(e,t,n){if(e!==null)if(e.memoizedProps!==t.pendingProps)nt=!0;else{if(!oc(e,n)&&(t.flags&128)===0)return nt=!1,T0(e,t,n);nt=(e.flags&131072)!==0}else nt=!1,Te&&(t.flags&1048576)!==0&&th(t,hi,t.index);switch(t.lanes=0,t.tag){case 16:e:{var s=t.pendingProps;if(e=Da(t.elementType),t.type=e,typeof e=="function")pr(e)?(s=Ia(e,s),t.tag=1,t=yp(null,t,e,s,n)):(t.tag=0,t=$r(null,t,e,s,n));else{if(e!=null){var i=e.$$typeof;if(i===F){t.tag=11,t=up(null,t,e,s,n);break e}else if(i===je){t.tag=14,t=hp(null,t,e,s,n);break e}else if(i===W){t.tag=10,t.type=e,t=Ap(null,t,n);break e}}throw t=ye(e)||e,Error(r(306,t,""))}}return t;case 0:return $r(e,t,t.type,t.pendingProps,n);case 1:return s=t.type,i=Ia(s,t.pendingProps),yp(e,t,s,i,n);case 3:e:{if(nn(t,t.stateNode.containerInfo),e===null)throw Error(r(387));s=t.pendingProps;var l=t.memoizedState;i=l.element,Tr(e,t),wi(t,s,null,n);var d=t.memoizedState;if(s=d.cache,Fn(t,et,s),s!==l.cache&&wr(t,[et],n,!0),xi(),s=d.element,l.isDehydrated)if(l={element:s,isDehydrated:!1,cache:d.cache},t.updateQueue.baseState=l,t.memoizedState=l,t.flags&256){t=xp(e,t,s,n);break e}else if(s!==i){i=Ft(Error(r(424)),t),pi(i),t=xp(e,t,s,n);break e}else for(e=t.stateNode.containerInfo,e.nodeType===9?e=e.body:e=e.nodeName==="HTML"?e.ownerDocument.body:e,Ye=en(e.firstChild),lt=t,Te=!0,Zn=null,Pt=!0,n=fh(t,null,s,n),t.child=n;n;)n.flags=n.flags&-3|134221824,n=n.sibling;else{if(Ea(),s===i){t=Dn(e,t,n);break e}it(e,t,s,n)}t=t.child}return t;case 26:return xs(e,t),e===null?(n=Jf(t.type,null,t.pendingProps,null))?t.memoizedState=n:Te||(t.stateNode=_f(t.type,t.pendingProps,Fe.current,t)):t.memoizedState=Jf(t.type,e.memoizedProps,t.pendingProps,e.memoizedState),null;case 27:return Js(t),e===null&&Te&&(s=t.stateNode=Qf(t.type,t.pendingProps,Fe.current),lt=t,Pt=!0,i=Ye,da(t.type)?(nd=i,Ye=en(s.firstChild)):Ye=i),it(e,t,t.pendingProps.children,n),xs(e,t),e===null&&(t.flags|=4194304),t.child;case 5:return e===null&&Te&&((i=s=Ye)&&(s=w1(s,t.type,t.pendingProps,Pt),s!==null?(t.stateNode=s,lt=t,Ye=en(s.firstChild),Pt=!1,i=!0):i=!1),i||Kn(t)),Js(t),i=t.type,l=t.pendingProps,d=e!==null?e.memoizedProps:null,s=l.children,Zc(i,l)?s=null:d!==null&&Zc(i,d)&&(t.flags|=32),t.memoizedState!==null&&(i=Mr(e,t,m0,null,null,n),Ls._currentValue=i),xs(e,t),it(e,t,s,n),t.child;case 6:return e===null&&Te&&((e=n=Ye)&&(n=b1(n,t.pendingProps,Pt),n!==null?(t.stateNode=n,lt=t,Ye=null,e=!0):e=!1),e||Kn(t)),null;case 13:return wp(e,t,n);case 4:return nn(t,t.stateNode.containerInfo),s=t.pendingProps,e===null?t.child=Ua(t,null,s,n):it(e,t,s,n),t.child;case 11:return up(e,t,t.type,t.pendingProps,n);case 7:return s=t.pendingProps,xs(e,t),it(e,t,s,n),t.child;case 8:return it(e,t,t.pendingProps.children,n),t.child;case 12:return it(e,t,t.pendingProps.children,n),t.child;case 10:return Ap(e,t,n);case 9:return i=t.type._context,s=t.pendingProps.children,Ra(t),i=pt(i),s=s(i),t.flags|=1,it(e,t,s,n),t.child;case 14:return hp(e,t,t.type,t.pendingProps,n);case 15:return pp(e,t,t.type,t.pendingProps,n);case 19:return ic(e,t,n);case 31:return k0(e,t,n);case 22:return fp(e,t,n,t.pendingProps);case 24:return Ra(t),s=pt(et),e===null?(i=Ar(),i===null&&(i=Ge,l=br(),i.pooledCache=l,l.refCount++,l!==null&&(i.pooledCacheLanes|=n),i=l),t.memoizedState={parent:s,cache:i},Cr(t),Fn(t,et,i)):((e.lanes&n)!==0&&(Tr(e,t),wi(t,null,null,n),xi()),i=e.memoizedState,l=t.memoizedState,i.parent!==s?(i={parent:s,cache:s},t.memoizedState=i,t.lanes===0&&(t.memoizedState=t.updateQueue.baseState=i),Fn(t,et,s)):(s=l.cache,Fn(t,et,s),s!==i.cache&&wr(t,[et],n,!0))),it(e,t,t.pendingProps.children,n),t.child;case 30:return t.stateNode===null&&(t.stateNode={autoName:null,paired:null,clones:null,ref:null}),s=t.pendingProps,s.name!=null&&s.name!=="auto"?t.flags|=e===null?18882560:18874368:Te&&ko(t),e!==null&&e.memoizedProps.name!==s.name?t.flags|=4194816:xs(e,t),it(e,t,s.children,n),t.child;case 29:throw t.pendingProps}throw Error(r(156,t.tag))}function Ln(e){e.flags|=4}function lc(e,t,n,s,i){var l;if((l=(e.mode&32)!==0)&&(l=n===null?em(t,s):em(t,s)&&(s.src!==n.src||s.srcSet!==n.srcSet)),l){if(e.flags|=16777216,(i&335544128)===i)if(e.stateNode.complete)e.flags|=8192;else if(of())e.flags|=8192;else throw La=Oo,kr}else e.flags&=-16777217}function Cp(e,t){if(t.type!=="stylesheet"||(t.state.loading&4)!==0)e.flags&=-16777217;else if(e.flags|=16777216,!tm(t))if(of())e.flags|=8192;else throw La=Oo,kr}function Jo(e,t){t!==null&&(e.flags|=4),e.flags&16384&&(t=e.tag!==22?$d():536870912,e.lanes|=t,ks|=t)}function Ti(e,t){if(!Te)switch(e.tailMode){case"visible":break;case"collapsed":for(var n=e.tail,s=null;n!==null;)n.alternate!==null&&(s=n),n=n.sibling;s===null?t||e.tail===null?e.tail=null:e.tail.sibling=null:s.sibling=null;break;default:for(t=e.tail,n=null;t!==null;)t.alternate!==null&&(n=t),t=t.sibling;n===null?e.tail=null:n.sibling=null}}function Xe(e){var t=e.alternate!==null&&e.alternate.child===e.child,n=0,s=0;if(t)for(var i=e.child;i!==null;)n|=i.lanes|i.childLanes,s|=i.subtreeFlags&1206910976,s|=i.flags&1206910976,i.return=e,i=i.sibling;else for(i=e.child;i!==null;)n|=i.lanes|i.childLanes,s|=i.subtreeFlags,s|=i.flags,i.return=e,i=i.sibling;return e.subtreeFlags|=s,e.childLanes=n,t}function S0(e,t,n){var s=t.pendingProps;switch(gr(t),t.tag){case 16:case 15:case 0:case 11:case 7:case 8:case 12:case 9:case 14:return Xe(t),null;case 1:return Xe(t),null;case 3:return n=t.stateNode,s=null,e!==null&&(s=e.memoizedState.cache),t.memoizedState.cache!==s&&(t.flags|=2048),Rn(et),Gn(),n.pendingContext&&(n.context=n.pendingContext,n.pendingContext=null),(e===null||e.child===null)&&(hs(t)?Ln(t):e===null||e.memoizedState.isDehydrated&&(t.flags&256)===0||(t.flags|=1024,yr())),Xe(t),null;case 26:var i=t.type,l=t.memoizedState;return e===null?(Ln(t),l!==null?(Xe(t),Cp(t,l)):(Xe(t),lc(t,i,null,s,n))):l?l!==e.memoizedState?(Ln(t),Xe(t),Cp(t,l)):(Xe(t),t.flags&=-16777217):(e=e.memoizedProps,e!==s&&Ln(t),Xe(t),lc(t,i,e,s,n)),null;case 27:if(ba(t),n=Fe.current,i=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==s&&Ln(t);else{if(!s){if(t.stateNode===null)throw Error(r(166));return Xe(t),t.subtreeFlags&=-33554433,null}e=st.current,hs(t)?ah(t):(e=Qf(i,s,n),t.stateNode=e,Ln(t))}return Xe(t),t.subtreeFlags&=-33554433,null;case 5:if(ba(t),i=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==s&&Ln(t);else{if(!s){if(t.stateNode===null)throw Error(r(166));return Xe(t),t.subtreeFlags&=-33554433,null}if(l=st.current,hs(t))ah(t);else{var d=Li(Fe.current);switch(l){case 1:l=d.createElementNS("http://www.w3.org/2000/svg",i);break;case 2:l=d.createElementNS("http://www.w3.org/1998/Math/MathML",i);break;default:switch(i){case"svg":l=d.createElementNS("http://www.w3.org/2000/svg",i);break;case"math":l=d.createElementNS("http://www.w3.org/1998/Math/MathML",i);break;case"script":l=d.createElement("div"),l.innerHTML="<script><\/script>",l=l.removeChild(l.firstChild);break;case"select":l=typeof s.is=="string"?d.createElement("select",{is:s.is}):d.createElement("select"),s.multiple?l.multiple=!0:s.size&&(l.size=s.size);break;default:l=typeof s.is=="string"?d.createElement(i,{is:s.is}):d.createElement(i)}}l[ht]=t,l[Tt]=s;e:for(d=t.child;d!==null;){if(d.tag===5||d.tag===6)l.appendChild(d.stateNode);else if(d.tag!==4&&d.tag!==27&&d.child!==null){d.child.return=d,d=d.child;continue}if(d===t)break e;for(;d.sibling===null;){if(d.return===null||d.return===t)break e;d=d.return}d.sibling.return=d.return,d=d.sibling}t.stateNode=l;e:switch(vt(l,i,s),i){case"button":case"input":case"select":case"textarea":s=!!s.autoFocus;break e;case"img":s=!0;break e;default:s=!1}s&&Ln(t)}}return Xe(t),t.subtreeFlags&=-33554433,lc(t,t.type,e===null?null:e.memoizedProps,t.pendingProps,n),null;case 6:if(e&&t.stateNode!=null)e.memoizedProps!==s&&Ln(t);else{if(typeof s!="string"&&t.stateNode===null)throw Error(r(166));if(e=Fe.current,hs(t)){if(e=t.stateNode,n=t.memoizedProps,s=null,i=lt,i!==null)switch(i.tag){case 27:case 5:s=i.memoizedProps}e[ht]=t,e=!!(e.nodeValue===n||s!==null&&s.suppressHydrationWarning===!0||Cf(e.nodeValue,n)),e||Kn(t,!0)}else e=Li(e).createTextNode(s),e[ht]=t,t.stateNode=e}return Xe(t),null;case 31:if(n=t.memoizedState,e===null||e.memoizedState!==null){if(s=hs(t),n!==null){if(e===null){if(!s)throw Error(r(318));if(e=t.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(r(557));e[ht]=t}else Ea(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;Xe(t),e=!1}else n=yr(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=n),e=!0;if(!e)return t.flags&256?(Ht(t),t):(Ht(t),null);if((t.flags&128)!==0)throw Error(r(558))}return Xe(t),null;case 13:if(s=t.memoizedState,e===null||e.memoizedState!==null&&e.memoizedState.dehydrated!==null){if(i=hs(t),s!==null&&s.dehydrated!==null){if(e===null){if(!i)throw Error(r(318));if(i=t.memoizedState,i=i!==null?i.dehydrated:null,!i)throw Error(r(317));i[ht]=t}else Ea(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;Xe(t),i=!1}else i=yr(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=i),i=!0;if(!i)return t.flags&256?(Ht(t),t):(Ht(t),null)}return Ht(t),(t.flags&128)!==0?(t.lanes=n,t):(n=s!==null,e=e!==null&&e.memoizedState!==null,n&&(s=t.child,i=null,s.alternate!==null&&s.alternate.memoizedState!==null&&s.alternate.memoizedState.cachePool!==null&&(i=s.alternate.memoizedState.cachePool.pool),l=null,s.memoizedState!==null&&s.memoizedState.cachePool!==null&&(l=s.memoizedState.cachePool.pool),l!==i&&(s.flags|=2048)),n!==e&&n&&(t.child.flags|=8192),Jo(t,t.updateQueue),Xe(t),null);case 4:return Gn(),e===null&&Bc(t.stateNode.containerInfo),t.flags|=67108864,Xe(t),null;case 10:return Rn(t.type),Xe(t),null;case 19:if(zr(t),s=t.memoizedState,s===null)return Xe(t),null;if(i=(t.flags&128)!==0,l=s.rendering,l===null)if(i)Ti(s,!1);else{if(Pe!==0||e!==null&&(e.flags&128)!==0)for(e=t.child;e!==null;){if(l=qo(e),l!==null){for(t.flags|=128,Ti(s,!1),e=l.updateQueue,t.updateQueue=e,Jo(t,e),t.subtreeFlags=0,e=n,n=t.child;n!==null;)Pu(n,e),n=n.sibling;return bi(t,mt.current&1|2),Te&&On(t,s.treeForkCount),t.child}e=e.sibling}s.tail!==null&&Mt()>rl&&(t.flags|=128,i=!0,Ti(s,!1),t.lanes=4194304)}else{if(!i)if(e=qo(l),e!==null){if(t.flags|=128,i=!0,e=e.updateQueue,t.updateQueue=e,Jo(t,e),Ti(s,!0),s.tail===null&&s.tailMode!=="collapsed"&&s.tailMode!=="visible"&&!l.alternate&&!Te)return Xe(t),null}else 2*Mt()-s.renderingStartTime>rl&&n!==536870912&&(t.flags|=128,i=!0,Ti(s,!1),t.lanes=4194304);s.isBackwards?(l.sibling=t.child,t.child=l):(e=s.last,e!==null?e.sibling=l:t.child=l,s.last=l)}if(s.tail!==null){e=s.tail;e:{for(n=e;n!==null;){if(n.alternate!==null){n=!1;break e}n=n.sibling}n=!0}return s.rendering=e,s.tail=e.sibling,s.renderingStartTime=Mt(),e.sibling=null,l=mt.current,l=i?l&1|2:l&1,s.tailMode==="visible"||s.tailMode==="collapsed"||!n||Te?bi(t,l):(n=l,we(ft,t),we(mt,n),xt===null&&(xt=t)),Te&&On(t,s.treeForkCount),e}return Xe(t),null;case 22:case 23:return Ht(t),Er(),s=t.memoizedState!==null,e!==null?e.memoizedState!==null!==s&&(t.flags|=8192):s&&(t.flags|=8192),s?(n&536870912)!==0&&(t.flags&128)===0&&(Xe(t),t.subtreeFlags&6&&(t.flags|=8192)):Xe(t),n=t.updateQueue,n!==null&&Jo(t,n.retryQueue),n=null,e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(n=e.memoizedState.cachePool.pool),s=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(s=t.memoizedState.cachePool.pool),s!==n&&(t.flags|=2048),e!==null&&Ie(qa),null;case 24:return n=null,e!==null&&(n=e.memoizedState.cache),t.memoizedState.cache!==n&&(t.flags|=2048),Rn(et),Xe(t),null;case 25:return null;case 30:return t.flags|=33554432,Xe(t),null}throw Error(r(156,t.tag))}function N0(e,t){switch(gr(t),t.tag){case 1:return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 3:return Rn(et),Gn(),e=t.flags,(e&65536)!==0&&(e&128)===0?(t.flags=e&-65537|128,t):null;case 26:case 27:case 5:return ba(t),null;case 31:if(t.memoizedState!==null){if(Ht(t),t.alternate===null)throw Error(r(340));Ea()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 13:if(Ht(t),e=t.memoizedState,e!==null&&e.dehydrated!==null){if(t.alternate===null)throw Error(r(340));Ea()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 19:return zr(t),e=t.flags,e&65536?(t.flags=e&-65537|128,e=t.memoizedState,e!==null&&(e.rendering=null,e.tail=null),t.flags|=4,t):null;case 4:return Gn(),null;case 10:return Rn(t.type),null;case 22:case 23:return Ht(t),Er(),e!==null&&Ie(qa),e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 24:return Rn(et),null;case 25:return null;default:return null}}function Tp(e,t){switch(gr(t),t.tag){case 3:Rn(et),Gn();break;case 26:case 27:case 5:ba(t);break;case 4:Gn();break;case 31:t.memoizedState!==null&&Ht(t);break;case 13:Ht(t);break;case 19:zr(t);break;case 10:Rn(t.type);break;case 22:case 23:Ht(t),Er(),e!==null&&Ie(qa);break;case 24:Rn(et)}}function Si(e,t){try{var n=t.updateQueue,s=n!==null?n.lastEffect:null;if(s!==null){var i=s.next;n=i;do{if((n.tag&e)===e){s=void 0;var l=n.create,d=n.inst;s=l(),d.destroy=s}n=n.next}while(n!==i)}}catch(p){Le(t,t.return,p)}}function aa(e,t,n){try{var s=t.updateQueue,i=s!==null?s.lastEffect:null;if(i!==null){var l=i.next;s=l;do{if((s.tag&e)===e){var d=s.inst,p=d.destroy;if(p!==void 0){d.destroy=void 0,i=t;var k=n,R=p;try{R()}catch(L){Le(i,k,L)}}}s=s.next}while(s!==l)}}catch(L){Le(t,t.return,L)}}function Sp(e){var t=e.updateQueue;if(t!==null){var n=e.stateNode;try{gh(t,n)}catch(s){Le(e,e.return,s)}}}function Np(e,t,n){n.props=Ia(e.type,e.memoizedProps),n.state=e.memoizedState;try{n.componentWillUnmount()}catch(s){Le(e,t,s)}}function gn(e,t){try{var n=e.ref;if(n!==null){switch(e.tag){case 26:case 27:case 5:var s=e.stateNode;break;case 30:var i=e.stateNode,l=Nn(e.memoizedProps,i);(i.ref===null||i.ref.name!==l)&&(i.ref=Df(l)),s=i.ref;break;case 7:if(e.stateNode===null){var d=new Xt(e);y(e.child,!1,y1,d,void 0,void 0),e.stateNode=d}s=e.stateNode;break;default:s=e.stateNode}typeof n=="function"?e.refCleanup=n(s):n.current=s}}catch(p){Le(e,t,p)}}function gt(e,t){var n=e.ref,s=e.refCleanup;if(n!==null)if(typeof s=="function")try{s()}catch(i){Le(e,t,i)}finally{e.refCleanup=null,e=e.alternate,e!=null&&(e.refCleanup=null)}else if(typeof n=="function")try{n(null)}catch(i){Le(e,t,i)}else n.current=null}function Wo(e,t){if((e.tag===5||e.tag===27||e.tag===6)&&e.alternate===null&&t!==null)for(var n=0;n<t.length;n++)Bf(e.stateNode,t[n])}function _p(e){for(var t=e.return;t!==null&&(cc(t)&&Bf(e.stateNode,t.stateNode),!rc(t));)t=t.return}function Ni(e){for(var t=e.return;t!==null&&(cc(t)&&x1(e.stateNode,t.stateNode),!rc(t));)t=t.return}function rc(e){return e.tag===5||e.tag===3||e.tag===27}function cc(e){return e&&e.tag===7&&e.stateNode!==null}function dc(e){var t=e.type,n=e.memoizedProps,s=e.stateNode;try{e:switch(t){case"button":case"input":case"select":case"textarea":n.autoFocus&&s.focus();break e;case"img":n.src?s.src=n.src:n.srcSet&&(s.srcset=n.srcSet)}}catch(i){Le(e,e.return,i)}}function uc(e,t,n){try{var s=e.stateNode;e1(s,e.type,n,t),s[Tt]=t}catch(i){Le(e,e.return,i)}}function Ep(e){return e.tag===5||e.tag===3||e.tag===26||e.tag===27&&da(e.type)||e.tag===4}function hc(e){e:for(;;){for(;e.sibling===null;){if(e.return===null||Ep(e.return))return null;e=e.return}for(e.sibling.return=e.return,e=e.sibling;e.tag!==5&&e.tag!==6&&e.tag!==18;){if(e.tag===27&&da(e.type)||e.flags&2||e.child===null||e.tag===4)continue e;e.child.return=e,e=e.child}if(!(e.flags&2))return e.stateNode}}function pc(e,t,n,s){var i=e.tag;if(i===5||i===6)i=e.stateNode,t?(n.nodeType===9?n.body:n.nodeName==="HTML"?n.ownerDocument.body:n).insertBefore(i,t):(t=n.nodeType===9?n.body:n.nodeName==="HTML"?n.ownerDocument.body:n,t.appendChild(i),n=n._reactRootContainer,n!=null||t.onclick!==null||(t.onclick=pn)),Wo(e,s),ze=!0;else if(i!==4&&(i===27&&(Wo(e,s),s=null,da(e.type)&&(n=e.stateNode,t=null)),e=e.child,e!==null))for(pc(e,t,n,s),e=e.sibling;e!==null;)pc(e,t,n,s),e=e.sibling}function Po(e,t,n,s){var i=e.tag;if(i===5||i===6)i=e.stateNode,t?n.insertBefore(i,t):n.appendChild(i),Wo(e,s),ze=!0;else if(i!==4&&(i===27&&(Wo(e,s),s=null,da(e.type)&&(n=e.stateNode)),e=e.child,e!==null))for(Po(e,t,n,s),e=e.sibling;e!==null;)Po(e,t,n,s),e=e.sibling}function Op(e){var t=e.stateNode,n=e.memoizedProps;try{for(var s=e.type,i=t.attributes;i.length;)t.removeAttributeNode(i[0]);vt(t,s,n),t[ht]=e,t[Tt]=n}catch(l){Le(e,e.return,l)}}var $o=!1,It=null;function zp(e){(e.tag===30||(e.subtreeFlags&33554432)!==0)&&($o=!0)}var vn=null;function Rp(){var e=vn;return vn=null,e}var Nt=0;function ws(e,t,n,s,i){return Nt=0,Mp(e.child,t,n,s,i)}function Mp(e,t,n,s,i){for(var l=!1;e!==null;){if(e.tag===5){var d=e.stateNode;if(s!==null){var p=Jc(d);s.push(p),p.view&&(l=!0)}else l||Jc(d).view&&(l=!0);$o=!0,Mf(d,Nt===0?t:t+"_"+Nt,n),Nt++}else(e.tag!==22||e.memoizedState===null)&&(e.tag===30&&i||Mp(e.child,t,n,s,i)&&(l=!0));e=e.sibling}return l}function yn(e,t){for(;e!==null;)e.tag===5?qf(e.stateNode,e.memoizedProps):(e.tag!==22||e.memoizedState===null)&&(e.tag===30&&t||yn(e.child,t)),e=e.sibling}function el(e){if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if((e.tag!==22||e.memoizedState===null)&&(el(e),e.tag===30&&(e.flags&18874368)!==0&&e.stateNode.paired)){var t=e.memoizedProps;if(t.name==null||t.name==="auto")throw Error(r(544));var n=t.name;t=_n(t.default,t.share),t!=="none"&&(ws(e,n,t,null,!1)||yn(e.child,!1))}e=e.sibling}}function fc(e,t){if(e.tag===30){var n=e.stateNode,s=e.memoizedProps,i=Nn(s,n),l=_n(s.default,n.paired?s.share:s.enter);l!=="none"?ws(e,i,l,null,!1)?(el(e),n.paired||t||Ns(e,s.onEnter)):yn(e.child,!1):el(e)}else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)fc(e,t),e=e.sibling;else el(e)}function mc(e){if(It!==null&&It.size!==0){var t=It;if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if(e.tag!==22||e.memoizedState===null){if(e.tag===30&&(e.flags&18874368)!==0){var n=e.memoizedProps,s=n.name;if(s!=null&&s!=="auto"){var i=t.get(s);if(i!==void 0){var l=_n(n.default,n.share);if(l!=="none"&&(ws(e,s,l,null,!1)?(l=e.stateNode,i.paired=l,l.paired=i,Ns(e,n.onShare)):yn(e.child,!1)),t.delete(s),t.size===0)break}}}mc(e)}e=e.sibling}}}function gc(e){if(e.tag===30){var t=e.memoizedProps,n=Nn(t,e.stateNode),s=It!==null?It.get(n):void 0,i=_n(t.default,s!==void 0?t.share:t.exit);i!=="none"&&(ws(e,n,i,null,!1)?s!==void 0?(i=e.stateNode,s.paired=i,i.paired=s,It.delete(n),Ns(e,t.onShare)):Ns(e,t.onExit):yn(e.child,!1)),It!==null&&mc(e)}else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)gc(e),e=e.sibling;else It!==null&&mc(e)}function qp(e){for(e=e.child;e!==null;){if(e.tag===30){var t=e.memoizedProps,n=Nn(t,e.stateNode);t=_n(t.default,t.update),e.flags&=-5,t!=="none"&&ws(e,n,t,e.memoizedState=[],!1)}else(e.subtreeFlags&33554432)!==0&&qp(e);e=e.sibling}}function vc(e){if((e.subtreeFlags&18874368)!==0)for(e=e.child;e!==null;){if(e.tag!==22||e.memoizedState===null){if(e.tag===30&&(e.flags&18874368)!==0){var t=e.stateNode;t.paired!==null&&(t.paired=null,yn(e.child,!1))}vc(e)}e=e.sibling}}function tl(e){if(e.tag===30)e.stateNode.paired=null,yn(e.child,!1),vc(e);else if((e.subtreeFlags&33554432)!==0)for(e=e.child;e!==null;)tl(e),e=e.sibling;else vc(e)}function Dp(e){for(e=e.child;e!==null;)e.tag===30?yn(e.child,!1):(e.subtreeFlags&33554432)!==0&&Dp(e),e=e.sibling}function yc(e,t,n,s,i,l,d){for(var p=!1;t!==null;){if(t.tag===5){var k=t.stateNode;if(l!==null&&Nt<l.length){var R=l[Nt],L=Jc(k);(R.view||L.view)&&(p=!0);var V;if(V=(e.flags&4)===0)if(L.clip)V=!0;else{V=R.rect;var O=L.rect;V=V.y!==O.y||V.x!==O.x||V.height!==O.height||V.width!==O.width}V&&(e.flags|=4),L.abs?L=!R.abs:(R=R.rect,L=L.rect,L=R.height!==L.height||R.width!==L.width),L&&(e.flags|=32)}else e.flags|=32;(e.flags&4)!==0&&Mf(k,Nt===0?n:n+"_"+Nt,i),p&&(e.flags&4)!==0||(vn===null&&(vn=[]),vn.push(k,Nt===0?s:s+"_"+Nt,t.memoizedProps)),Nt++}else(t.tag!==22||t.memoizedState===null)&&(t.tag===30&&d?e.flags|=t.flags&32:yc(e,t.child,n,s,i,l,d)&&(p=!0));t=t.sibling}return p}function Lp(e,t){for(e=e.child;e!==null;){if(e.tag===30){var n=e.memoizedProps,s=e.stateNode,i=Nn(n,s),l=_n(n.default,n.update),d;d=e.memoizedState,e.memoizedState=null,s=e;var p=e.child;Nt=0,i=yc(s,p,i,i,l,d,!1),(e.flags&4)!==0&&i&&Ns(e,n.onUpdate)}else(e.subtreeFlags&33554432)!==0&&Lp(e);e=e.sibling}}var rt=!1,qe=!1,xn=!1,xc=!1,Up=typeof WeakSet=="function"?WeakSet:Set,ct=null,wn=!1,_i=!1,nl=!1,wc=!1;function _0(e,t,n){if(e=e.containerInfo,Vc=Us,e=Bu(e),or(e)){if("selectionStart"in e)var s={start:e.selectionStart,end:e.selectionEnd};else e:{s=(s=e.ownerDocument)&&s.defaultView||window;var i=s.getSelection&&s.getSelection();if(i&&i.rangeCount!==0){s=i.anchorNode;var l=i.anchorOffset,d=i.focusNode;i=i.focusOffset;try{s.nodeType,d.nodeType}catch{s=null;break e}var p=0,k=-1,R=-1,L=0,V=0,O=e,q=null;t:for(;;){for(var P;O!==s||l!==0&&O.nodeType!==3||(k=p+l),O!==d||i!==0&&O.nodeType!==3||(R=p+i),O.nodeType===3&&(p+=O.nodeValue.length),(P=O.firstChild)!==null;)q=O,O=P;for(;;){if(O===e)break t;if(q===s&&++L===l&&(k=p),q===d&&++V===i&&(R=p),(P=O.nextSibling)!==null)break;O=q,q=O.parentNode}O=P}s=k===-1||R===-1?null:{start:k,end:R}}else s=null}s=s||{start:0,end:0}}else s=null;for(Qc={focusedElem:e,selectionRange:s},Us=!1,n=(n&335544064)===n,ct=t,t=n?9270:1024;ct!==null;){if(e=ct,n&&(s=e.deletions,s!==null))for(l=0;l<s.length;l++)n&&gc(s[l]);if(e.alternate===null&&(e.flags&2)!==0)n&&zp(e),al(n);else{if(e.tag===22){if(s=e.alternate,e.memoizedState!==null){s!==null&&s.memoizedState===null&&n&&gc(s),al(n);continue}else if(s!==null&&s.memoizedState!==null){n&&zp(e),al(n);continue}}s=e.child,(e.subtreeFlags&t)!==0&&s!==null?(s.return=e,ct=s):(n&&qp(e),al(n))}}It=null}function al(e){for(;ct!==null;){var t=ct,n=e,s=t.alternate,i=t.flags;switch(t.tag){case 0:case 11:case 15:break;case 1:if((i&1024)!==0&&s!==null){n=void 0,i=s.memoizedProps,s=s.memoizedState;var l=t.stateNode;try{var d=Ia(t.type,i);n=l.getSnapshotBeforeUpdate(d,s),l.__reactInternalSnapshotBeforeUpdate=n}catch(p){Le(t,t.return,p)}}break;case 3:if((i&1024)!==0){if(s=t.stateNode.containerInfo,n=s.nodeType,n===9)$c(s);else if(n===1)switch(s.nodeName){case"HEAD":case"HTML":case"BODY":$c(s);break;default:s.textContent=""}}break;case 5:case 26:case 27:case 6:case 4:case 17:break;case 30:n&&s!==null&&(n=Nn(s.memoizedProps,s.stateNode),i=t.memoizedProps,i=_n(i.default,i.update),i!=="none"&&ws(s,n,i,s.memoizedState=[],!0));break;default:if((i&1024)!==0)throw Error(r(163))}if(s=t.sibling,s!==null){s.return=t.return,ct=s;break}ct=t.return}}function Hp(e,t,n){var s=n.flags;switch(n.tag){case 0:case 11:case 15:bn(e,n),s&4&&Si(5,n);break;case 1:if(bn(e,n),s&4)if(e=n.stateNode,t===null)try{e.componentDidMount()}catch(d){Le(n,n.return,d)}else{var i=Ia(n.type,t.memoizedProps);t=t.memoizedState;try{e.componentDidUpdate(i,t,e.__reactInternalSnapshotBeforeUpdate)}catch(d){Le(n,n.return,d)}}s&64&&Sp(n),s&512&&gn(n,n.return);break;case 3:if(bn(e,n),s&64&&(e=n.updateQueue,e!==null)){if(t=null,n.child!==null)switch(n.child.tag){case 27:case 5:t=n.child.stateNode;break;case 1:t=n.child.stateNode}try{gh(e,t)}catch(d){Le(n,n.return,d)}}break;case 27:t===null&&s&4&&Op(n);case 26:case 5:bn(e,n),t===null&&s&4&&dc(n),s&512&&gn(n,n.return);break;case 12:bn(e,n);break;case 31:bn(e,n),s&4&&Yp(e,n);break;case 13:bn(e,n),s&4&&Xp(e,n),s&64&&(e=n.memoizedState,e!==null&&(e=e.dehydrated,e!==null&&(n=G0.bind(null,n),j1(e,n))));break;case 22:if(s=n.memoizedState!==null||rt,!s){var l=t!==null&&t.memoizedState!==null||qe;t=rt,i=qe,rt=s,(qe=l)&&!i?(s=2,(n.subtreeFlags&8772)!==0&&(s|=1),ln(e,n,s)):bn(e,n),rt=t,qe=i}break;case 30:bn(e,n),s&512&&gn(n,n.return);break;case 7:s&512&&gn(n,n.return);default:bn(e,n)}}function bc(e,t){for(e=e.child;e!==null;)Ip(e,t),e=e.sibling}function Ip(e,t){switch(e.tag){case 5:case 26:try{var n=e.stateNode;if(t){var s=n.style;typeof s.setProperty=="function"?s.setProperty("display","none","important"):s.display="none"}else{var i=e.stateNode,l=e.memoizedProps.style,d=l!=null&&l.hasOwnProperty("display")?l.display:null;i.style.display=d==null||typeof d=="boolean"?"":(""+d).trim()}}catch(k){Le(e,e.return,k)}jc(e,t);break;case 6:try{e.stateNode.nodeValue=t?"":e.memoizedProps,ze=!0}catch(k){Le(e,e.return,k)}break;case 18:try{var p=e.stateNode;t?Rf(p,!0):Rf(e.stateNode,!1)}catch(k){Le(e,e.return,k)}break;case 22:case 23:e.memoizedState===null&&bc(e,t);break;default:bc(e,t)}}function jc(e,t){if(e.subtreeFlags&67108864)for(e=e.child;e!==null;){e:{var n=e,s=t;switch(n.tag){case 4:Ip(n,s);break e;case 22:n.memoizedState===null&&jc(n,s);break e;default:jc(n,s)}}e=e.sibling}}function Gp(e){var t=e.alternate;t!==null&&(e.alternate=null,Gp(t)),e.child=null,e.deletions=null,e.sibling=null,e.tag===5&&(t=e.stateNode,t!==null&&lo(t)),e.stateNode=null,e.return=null,e.dependencies=null,e.memoizedProps=null,e.memoizedState=null,e.pendingProps=null,e.stateNode=null,e.updateQueue=null}var Ke=null,_t=!1;function sn(e,t,n){for(n=n.child;n!==null;)Bp(e,t,n),n=n.sibling}function Bp(e,t,n){if(qt&&typeof qt.onCommitFiberUnmount=="function")try{qt.onCommitFiberUnmount(Ps,n)}catch{}switch(n.tag){case 26:qe||gt(n,t),sn(e,t,n),n.memoizedState?n.memoizedState.count--:n.stateNode&&!qe&&(n=n.stateNode,n.parentNode.removeChild(n));break;case 27:qe||gt(n,t),Ni(n);var s=Ke,i=_t;da(n.type)&&(Ke=n.stateNode,_t=!1),sn(e,t,n),Zf(n.stateNode,n.type,n.memoizedProps),Ke=s,_t=i;break;case 5:qe||gt(n,t),Ni(n);case 6:if(n.tag===6&&Ni(n),s=Ke,i=_t,Ke=null,sn(e,t,n),Ke=s,_t=i,Ke!==null)if(_t)try{(Ke.nodeType===9?Ke.body:Ke.nodeName==="HTML"?Ke.ownerDocument.body:Ke).removeChild(n.stateNode),ze=!0}catch(l){Le(n,t,l)}else try{Ke.removeChild(n.stateNode),ze=!0}catch(l){Le(n,t,l)}break;case 18:Ke!==null&&(_t?(e=Ke,zf(e.nodeType===9?e.body:e.nodeName==="HTML"?e.ownerDocument.body:e,n.stateNode),Hs(e)):zf(Ke,n.stateNode));break;case 4:s=Ke,i=_t,Ke=n.stateNode.containerInfo,_t=!0,sn(e,t,n),Ke=s,_t=i;break;case 0:case 11:case 14:case 15:aa(2,n,t),qe||aa(4,n,t),sn(e,t,n);break;case 1:qe||(gt(n,t),s=n.stateNode,typeof s.componentWillUnmount=="function"&&Np(n,t,s)),sn(e,t,n);break;case 21:sn(e,t,n);break;case 22:qe=(s=qe)||n.memoizedState!==null,sn(e,t,n),qe=s;break;case 30:gt(n,t),sn(e,t,n);break;case 7:qe||gt(n,t),sn(e,t,n);break;default:sn(e,t,n)}}function Yp(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null))){e=e.dehydrated;try{Hs(e)}catch(n){Le(t,t.return,n)}}}function Xp(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null&&(e=e.dehydrated,e!==null))))try{Hs(e)}catch(n){Le(t,t.return,n)}}function E0(e){switch(e.tag){case 31:case 13:case 19:var t=e.stateNode;return t===null&&(t=e.stateNode=new Up),t;case 22:return e=e.stateNode,t=e._retryCache,t===null&&(t=e._retryCache=new Up),t;default:throw Error(r(435,e.tag))}}function sl(e,t){var n=E0(e);t.forEach(function(s){if(!n.has(s)){n.add(s);var i=B0.bind(null,e,s);s.then(i,i)}})}function At(e,t,n){var s=t.deletions;if(s!==null)for(var i=0;i<s.length;i++){var l=s[i],d=e,p=t,k=p;e:for(;k!==null;){switch(k.tag){case 27:if(da(k.type)){Ke=k.stateNode,_t=!1;break e}break;case 5:Ke=k.stateNode,_t=!1;break e;case 3:case 4:Ke=k.stateNode.containerInfo,_t=!0;break e}k=k.return}if(Ke===null)throw Error(r(160));Bp(d,p,l),Ke=null,_t=!1,d=l.alternate,d!==null&&(d.return=null),l.return=null}if(t.subtreeFlags&13886)for(t=t.child;t!==null;)Vp(t,e,n),t=t.sibling}var on=null;function Vp(e,t,n){var s=e.alternate,i=e.flags;switch(e.tag){case 0:case 11:case 14:case 15:if(i&4&&(s=e.updateQueue,s=s!==null?s.events:null,s!==null))for(var l=0;l<s.length;l++){var d=s[l];d.ref.impl=d.nextImpl}At(t,e,n),kt(e),i&4&&(aa(3,e,e.return),Si(3,e),aa(5,e,e.return));break;case 1:At(t,e,n),kt(e),i&512&&(qe||s===null||gt(s,s.return)),i&64&&rt&&(e=e.updateQueue,e!==null&&(t=e.callbacks,t!==null&&(n=e.shared.hiddenCallbacks,e.shared.hiddenCallbacks=n===null?t:n.concat(t))));break;case 26:if(l=on,At(t,e,n),kt(e),i&512&&(qe||s===null||gt(s,s.return)),i&4)if(i=s!==null?s.memoizedState:null,n=e.memoizedState,s===null)if(n===null)if(e.stateNode===null)if(rt)e.stateNode=_f(e.type,e.memoizedProps,t.containerInfo,e);else{e:{t=e.type,n=e.memoizedProps,i=l.ownerDocument||l;t:switch(t){case"title":s=i.getElementsByTagName("title")[0],(!s||s[ti]||s[ht]||s.namespaceURI==="http://www.w3.org/2000/svg"||s.hasAttribute("itemprop"))&&(s=i.createElement(t),i.head.insertBefore(s,i.querySelector("head > title"))),vt(s,t,n),s[ht]=e,ot(s),t=s;break e;case"link":if(l=$f("link","href",i).get(t+(n.href||""))){for(d=0;d<l.length;d++)if(s=l[d],s.getAttribute("href")===(n.href==null||n.href===""?null:n.href)&&s.getAttribute("rel")===(n.rel==null?null:n.rel)&&s.getAttribute("title")===(n.title==null?null:n.title)&&s.getAttribute("crossorigin")===(n.crossOrigin==null?null:n.crossOrigin)){l.splice(d,1);break t}}s=i.createElement(t),vt(s,t,n),i.head.appendChild(s);break;case"meta":if(l=$f("meta","content",i).get(t+(n.content||""))){for(d=0;d<l.length;d++)if(s=l[d],s.getAttribute("content")===(n.content==null?null:""+n.content)&&s.getAttribute("name")===(n.name==null?null:n.name)&&s.getAttribute("property")===(n.property==null?null:n.property)&&s.getAttribute("http-equiv")===(n.httpEquiv==null?null:n.httpEquiv)&&s.getAttribute("charset")===(n.charSet==null?null:n.charSet)){l.splice(d,1);break t}}s=i.createElement(t),vt(s,t,n),i.head.appendChild(s);break;default:throw Error(r(468,t))}s[ht]=e,ot(s),t=s}e.stateNode=t}else rt||od(l,e.type,e.stateNode);else e.stateNode=Pf(l,n,e.memoizedProps);else i!==n?(i===null?(t=s.stateNode,t===null||qe||t.parentNode.removeChild(t)):i.count--,n===null?rt||od(l,e.type,e.stateNode):Pf(l,n,e.memoizedProps)):n===null&&e.stateNode!==null&&uc(e,e.memoizedProps,s.memoizedProps);break;case 27:At(t,e,n),kt(e),i&512&&(qe||s===null||gt(s,s.return)),s!==null&&i&4&&uc(e,e.memoizedProps,s.memoizedProps);break;case 5:if(l=xn,xn=!1,At(t,e,n),xn=l,kt(e),i&512&&(qe||s===null||gt(s,s.return)),e.flags&32){t=e.stateNode;try{ns(t,""),ze=!0}catch(L){Le(e,e.return,L)}}i&4&&e.stateNode!=null&&(t=e.memoizedProps,uc(e,t,s!==null?s.memoizedProps:t)),i&1024&&(xc=!0);break;case 6:if(At(t,e,n),kt(e),i&4){if(e.stateNode===null)throw Error(r(162));t=e.memoizedProps,n=e.stateNode;try{n.nodeValue=t,ze=!0}catch(L){Le(e,e.return,L)}}break;case 3:if(ze=!1,xl=null,l=on,on=Ui(t.containerInfo),At(t,e,n),on=l,kt(e),i&4&&s!==null&&s.memoizedState.isDehydrated)try{Hs(t.containerInfo)}catch(L){Le(e,e.return,L)}xc&&(xc=!1,Qp(e)),ze=!1;break;case 4:i=xn,xn=rt,s=hu(),l=on,on=Ui(e.stateNode.containerInfo),At(t,e,n),kt(e),on=l,ze&&_i&&(nl=!0),ze=s,xn=i;break;case 12:At(t,e,n),kt(e);break;case 31:At(t,e,n),kt(e),i&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,sl(e,t)));break;case 13:At(t,e,n),kt(e),e.child.flags&8192&&e.memoizedState!==null!=(s!==null&&s.memoizedState!==null)&&(ll=Mt()),i&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,sl(e,t)));break;case 22:l=e.memoizedState!==null,d=s!==null&&s.memoizedState!==null;var p=rt,k=qe,R=xn;rt=p||l,xn=R||l,qe=k||d,At(t,e,n),qe=k,xn=R,rt=p,kt(e),i&8192&&(t=e.stateNode,t._visibility=l?t._visibility&-2:t._visibility|1,!l||s===null||d||rt||qe||(t=d||qe,n=rt,s=qe,rt=l||rt,qe=t,sa(e,2),rt=n,qe=s),!l&&xn||bc(e,l)),i&4&&(t=e.updateQueue,t!==null&&(n=t.retryQueue,n!==null&&(t.retryQueue=null,sl(e,n))));break;case 19:At(t,e,n),kt(e),i&4&&(t=e.updateQueue,t!==null&&(e.updateQueue=null,sl(e,t)));break;case 30:i&512&&(qe||s===null||gt(s,s.return)),i=hu(),l=_i,d=(n&335544064)===n,p=e.memoizedProps,_i=d&&_n(p.default,p.update)!=="none",At(t,e,n),kt(e),d&&s!==null&&ze&&(e.flags|=4),_i=l,ze=i;break;case 21:break;case 7:i&512&&(qe||s===null||gt(s,s.return)),s&&s.stateNode!==null&&(s.stateNode._fragmentFiber=e);default:At(t,e,n),kt(e)}}function kt(e){var t=e.flags;if(t&2){try{for(var n,s=e.return;s!==null;){if(Ep(s)){n=s;break}s=s.return}s=null;for(var i=e.return;i!==null;){if(cc(i)){var l=i.stateNode;s===null?s=[l]:s.push(l)}if(rc(i))break;i=i.return}var d=s;if(n==null)throw Error(r(160));switch(n.tag){case 27:var p=n.stateNode,k=hc(e);Po(e,k,p,d);break;case 5:var R=n.stateNode;n.flags&32&&(ns(R,""),n.flags&=-33);var L=hc(e);Po(e,L,R,d);break;case 3:case 4:var V=n.stateNode.containerInfo,O=hc(e);pc(e,O,V,d);break;default:throw Error(r(161))}}catch(q){Le(e,e.return,q)}e.flags&=-3}t&4096&&(e.flags&=-4097)}function Qp(e){if(e.subtreeFlags&1024)for(e=e.child;e!==null;){var t=e;Qp(t),t.tag===5&&t.flags&1024&&(t=t.stateNode,Us=!0,t.reset(),Us=!1),e=e.sibling}}function bs(e,t){if(t.subtreeFlags&9270)for(t=t.child;t!==null;)Zp(t,e),t=t.sibling;else Lp(t)}function Zp(e,t){var n=e.alternate;if(n===null)fc(e,!1);else switch(e.tag){case 3:if(wc=wn=!1,Rp(),bs(t,e),!wn&&!nl){if(e=vn,e!==null)for(var s=0;s<e.length;s+=3){n=e[s];var i=e[s+1];qf(n,e[s+2]),n=n.ownerDocument.documentElement,n!==null&&n.animate({opacity:[0,0],pointerEvents:["none","none"]},{duration:0,fill:"forwards",pseudoElement:"::view-transition-group("+i+")"})}e=t.containerInfo,e=e.nodeType===9?e.documentElement:e.ownerDocument.documentElement,e!==null&&e.style.viewTransitionName===""&&(e.style.viewTransitionName="none",e.animate({opacity:[0,0],pointerEvents:["none","none"]},{duration:0,fill:"forwards",pseudoElement:"::view-transition-group(root)"}),e.animate({width:[0,0],height:[0,0]},{duration:0,fill:"forwards",pseudoElement:"::view-transition"})),wc=!0}vn=null;break;case 5:bs(t,e);break;case 4:s=wn,wn=!1,bs(t,e),wn&&(nl=!0),wn=s;break;case 22:e.memoizedState===null&&(n.memoizedState!==null?fc(e,!1):bs(t,e));break;case 30:s=wn,i=Rp(),wn=!1,bs(t,e),wn&&(e.flags|=4);var l=e.memoizedProps,d=e.stateNode;t=Nn(l,d),d=Nn(n.memoizedProps,d);var p=_n(l.default,l.update);p==="none"?t=!1:(l=n.memoizedState,n.memoizedState=null,n=e.child,Nt=0,t=yc(e,n,t,d,p,l,!0),Nt!==(l===null?0:l.length)&&(e.flags|=32)),(e.flags&4)!==0&&t?(Ns(e,e.memoizedProps.onUpdate),vn=i):i!==null&&(i.push.apply(i,vn),vn=i),wn=(e.flags&32)!==0?!0:s;break;default:bs(t,e)}}function bn(e,t){if(t.subtreeFlags&8772)for(t=t.child;t!==null;)Hp(e,t.alternate,t),t=t.sibling}function sa(e,t){for(e=e.child;e!==null;){var n=e,s=t;switch(n.tag){case 0:case 11:case 14:case 15:aa(4,n,n.return),sa(n,s);break;case 1:gt(n,n.return);var i=n.stateNode;typeof i.componentWillUnmount=="function"&&Np(n,n.return,i),sa(n,s);break;case 27:(s&2)!==0&&Zf(n.stateNode,n.type,n.memoizedProps);case 5:gt(n,n.return),n.tag!==5&&n.tag!==27||Ni(n),sa(n,s);break;case 6:Ni(n);break;case 26:gt(n,n.return),i=n.stateNode,n.memoizedState!==null||i===null||qe||i.parentNode.removeChild(i),sa(n,s);break;case 22:n.memoizedState===null&&sa(n,s);break;case 30:gt(n,n.return),sa(n,s);break;case 7:gt(n,n.return);default:sa(n,s)}e=e.sibling}}function ln(e,t,n){for(n=(t.subtreeFlags&8772)!==0?n:n&-2,t=t.child;t!==null;){var s=t.alternate,i=e,l=t,d=l.flags,p=(n&1)!==0;switch(l.tag){case 0:case 11:case 15:ln(i,l,n),Si(4,l);break;case 1:if(ln(i,l,n),s=l,i=s.stateNode,typeof i.componentDidMount=="function")try{i.componentDidMount()}catch(L){Le(s,s.return,L)}if(s=l,i=s.updateQueue,i!==null){var k=s.stateNode;try{var R=i.shared.hiddenCallbacks;if(R!==null)for(i.shared.hiddenCallbacks=null,i=0;i<R.length;i++)mh(R[i],k)}catch(L){Le(s,s.return,L)}}p&&d&64&&Sp(l),gn(l,l.return);break;case 27:(n&2)!==0&&Op(l);case 5:l.tag!==5&&l.tag!==27||_p(l),ln(i,l,n),p&&s===null&&d&4&&dc(l),gn(l,l.return);break;case 6:_p(l);break;case 26:k=l.stateNode,l.memoizedState!==null||k===null||rt||od(Ui(k.ownerDocument),l.type,k),ln(i,l,n),p&&s===null&&d&4&&dc(l),gn(l,l.return);break;case 12:ln(i,l,n);break;case 31:ln(i,l,n),p&&d&4&&Yp(i,l);break;case 13:ln(i,l,n),p&&d&4&&Xp(i,l);break;case 22:l.memoizedState===null&&ln(i,l,n),gn(l,l.return);break;case 30:ln(i,l,n),gn(l,l.return);break;case 7:gn(l,l.return);default:ln(i,l,n)}t=t.sibling}}function Ac(e,t){var n=null;e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(n=e.memoizedState.cachePool.pool),e=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(e=t.memoizedState.cachePool.pool),e!==n&&(e!=null&&e.refCount++,n!=null&&fi(n))}function kc(e,t){e=null,t.alternate!==null&&(e=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==e&&(t.refCount++,e!=null&&fi(e))}function $t(e,t,n,s){var i=(n&335544064)===n;if(t.subtreeFlags&(i?10262:10256))for(t=t.child;t!==null;)Kp(e,t,n,s),t=t.sibling;else i&&Dp(t)}function Kp(e,t,n,s){var i=(n&335544064)===n;i&&t.alternate===null&&t.return!==null&&t.return.alternate!==null&&tl(t);var l=t.flags;switch(t.tag){case 0:case 11:case 15:$t(e,t,n,s),l&2048&&Si(9,t);break;case 1:$t(e,t,n,s);break;case 3:$t(e,t,n,s),i&&wc&&(e=e.containerInfo,e=e.nodeType===9?e.body:e.nodeName==="HTML"?e.ownerDocument.body:e,e.style.viewTransitionName==="root"&&(e.style.viewTransitionName=""),e=e.ownerDocument.documentElement,e!==null&&e.style.viewTransitionName==="none"&&(e.style.viewTransitionName="")),l&2048&&(l=null,t.alternate!==null&&(l=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==l&&(t.refCount++,l!=null&&fi(l)));break;case 12:if(l&2048){$t(e,t,n,s),l=t.stateNode;try{var d=t.memoizedProps,p=d.id,k=d.onPostCommit;typeof k=="function"&&k(p,t.alternate===null?"mount":"update",l.passiveEffectDuration,-0)}catch(R){Le(t,t.return,R)}}else $t(e,t,n,s);break;case 31:$t(e,t,n,s);break;case 13:$t(e,t,n,s);break;case 23:break;case 22:d=t.stateNode,p=t.alternate,t.memoizedState!==null?(i&&p!==null&&p.memoizedState===null&&tl(p),d._visibility&2?$t(e,t,n,s):Ei(e,t)):(i&&p!==null&&p.memoizedState!==null&&tl(t),d._visibility&2?$t(e,t,n,s):(d._visibility|=2,js(e,t,n,s,(t.subtreeFlags&10256)!==0||!1))),l&2048&&Ac(p,t);break;case 24:$t(e,t,n,s),l&2048&&kc(t.alternate,t);break;case 30:i&&(l=t.alternate,l!==null&&(yn(l.child,!0),yn(t.child,!0))),$t(e,t,n,s);break;default:$t(e,t,n,s)}}function js(e,t,n,s,i){for(i=i&&((t.subtreeFlags&10256)!==0||!1),t=t.child;t!==null;){var l=e,d=t,p=n,k=s,R=d.flags;switch(d.tag){case 0:case 11:case 15:js(l,d,p,k,i),Si(8,d);break;case 23:break;case 22:var L=d.stateNode;d.memoizedState!==null?L._visibility&2?js(l,d,p,k,i):Ei(l,d):(L._visibility|=2,js(l,d,p,k,i)),i&&R&2048&&Ac(d.alternate,d);break;case 24:js(l,d,p,k,i),i&&R&2048&&kc(d.alternate,d);break;default:js(l,d,p,k,i)}t=t.sibling}}function Ei(e,t){if(t.subtreeFlags&10256)for(t=t.child;t!==null;){var n=e,s=t,i=s.flags;switch(s.tag){case 22:Ei(n,s),i&2048&&Ac(s.alternate,s);break;case 24:Ei(n,s),i&2048&&kc(s.alternate,s);break;default:Ei(n,s)}t=t.sibling}}var Ga=8192;function Ba(e,t,n){if(e.subtreeFlags&Ga)for(e=e.child;e!==null;)Fp(e,t,n),e=e.sibling}function Fp(e,t,n){switch(e.tag){case 26:Ba(e,t,n),e.flags&Ga&&(e.memoizedState!==null?D1(n,on,e.memoizedState,e.memoizedProps):(e=e.stateNode,(t&335544128)===t&&am(n,e)));break;case 5:Ba(e,t,n),e.flags&Ga&&(e=e.stateNode,(t&335544128)===t&&am(n,e));break;case 3:case 4:var s=on;on=Ui(e.stateNode.containerInfo),Ba(e,t,n),on=s;break;case 22:e.memoizedState===null&&(s=e.alternate,s!==null&&s.memoizedState!==null?(s=Ga,Ga=16777216,Ba(e,t,n),Ga=s):Ba(e,t,n));break;case 30:if((e.flags&Ga)!==0&&(s=e.memoizedProps.name,s!=null&&s!=="auto")){var i=e.stateNode;i.paired=null,It===null&&(It=new Map),It.set(s,i)}Ba(e,t,n);break;default:Ba(e,t,n)}}function Jp(e){var t=e.alternate;if(t!==null&&(e=t.child,e!==null)){t.child=null;do t=e.sibling,e.sibling=null,e=t;while(e!==null)}}function Oi(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var n=0;n<t.length;n++){var s=t[n];ct=s,Pp(s,e)}Jp(e)}if(e.subtreeFlags&10256)for(e=e.child;e!==null;)Wp(e),e=e.sibling}function Wp(e){switch(e.tag){case 0:case 11:case 15:Oi(e),e.flags&2048&&aa(9,e,e.return);break;case 3:Oi(e);break;case 12:Oi(e);break;case 22:var t=e.stateNode;e.memoizedState!==null&&t._visibility&2&&(e.return===null||e.return.tag!==13)?(t._visibility&=-3,il(e)):Oi(e);break;default:Oi(e)}}function il(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var n=0;n<t.length;n++){var s=t[n];ct=s,Pp(s,e)}Jp(e)}for(e=e.child;e!==null;){switch(t=e,t.tag){case 0:case 11:case 15:aa(8,t,t.return),il(t);break;case 22:n=t.stateNode,n._visibility&2&&(n._visibility&=-3,il(t));break;default:il(t)}e=e.sibling}}function Pp(e,t){for(;ct!==null;){var n=ct;switch(n.tag){case 0:case 11:case 15:aa(8,n,t);break;case 23:case 22:if(n.memoizedState!==null&&n.memoizedState.cachePool!==null){var s=n.memoizedState.cachePool.pool;s!=null&&s.refCount++}break;case 24:fi(n.memoizedState.cache)}if(s=n.child,s!==null)s.return=n,ct=s;else e:for(n=e;ct!==null;){s=ct;var i=s.sibling,l=s.return;if(Gp(s),s===n){ct=null;break e}if(i!==null){i.return=l,ct=i;break e}ct=l}}}var O0={getCacheForType:function(e){var t=pt(et),n=t.data.get(e);return n===void 0&&(n=e(),t.data.set(e,n)),n},cacheSignal:function(){return pt(et).controller.signal}},z0=typeof WeakMap=="function"?WeakMap:Map,Me=0,Ge=null,Se=null,_e=0,De=0,Gt=null,ia=!1,As=!1,Cc=!1,Un=0,Pe=0,oa=0,Ya=0,ol=0,Bt=0,ks=0,zi=null,Et=null,Tc=!1,ll=0,$p=0,rl=1/0,cl=null,la=null,Je=0,rn=null,Xa=null,jn=0,Sc=0,Nc=null,ef=null,Cs=null,Ts=null,Ss=null,Ri=0,dl=null;function Yt(){return(Me&2)!==0&&_e!==0?_e&-_e:ae.T!==null?Uc():au()}function tf(){if(Bt===0)if((_e&536870912)===0||Te){var e=ao;ao<<=1,(ao&3932160)===0&&(ao=262144),Bt=e}else Bt=536870912;return e=ft.current,e!==null&&(e.flags|=32),Bt}function Ns(e,t){if(t!=null){var n=e.stateNode,s=n.ref;s===null&&(s=n.ref=Df(Nn(e.memoizedProps,n))),Ts===null&&(Ts=[]),Ts.push(t.bind(null,s))}}function Ot(e,t,n){(e===Ge&&(De===2||De===9)||e.cancelPendingCommit!==null)&&(_s(e,0),ra(e,_e,Bt,!1)),ei(e,n),((Me&2)===0||e!==Ge)&&(e===Ge&&((Me&2)===0&&(Ya|=n),Pe===4&&ra(e,_e,Bt,!1)),An(e))}function nf(e,t,n){if((Me&6)!==0)throw Error(r(327));var s=!n&&(t&127)===0&&(t&e.expiredLanes)===0||$s(e,t),i=s?q0(e,t):Ec(e,t,!0),l=s;do{if(i===0){As&&!s&&ra(e,t,0,!1);break}else{if(n=e.current.alternate,l&&!R0(n)){i=Ec(e,t,!1),l=!1;continue}if(i===2){if(l=t,e.errorRecoveryDisabledLanes&l)var d=0;else d=e.pendingLanes&-536870913,d=d!==0?d:d&536870912?536870912:0;if(d!==0){t=d;e:{var p=e;i=zi;var k=p.current.memoizedState.isDehydrated;if(k&&(_s(p,d).flags|=256),d=Ec(p,d,!1),d!==2&&d!==6){if(Cc&&!k){p.errorRecoveryDisabledLanes|=l,Ya|=l,i=4;break e}l=Et,Et=i,l!==null&&(Et===null?Et=l:Et.push.apply(Et,l))}i=d}if(l=!1,i!==2)continue}}if(i===1){_s(e,0),ra(e,t,0,!0);break}e:{switch(s=e,l=i,l){case 0:case 1:throw Error(r(345));case 4:if((t&4194048)!==t&&(t&62914560)!==t)break;case 6:ra(s,t,Bt,!ia);break e;case 2:Et=null;break;case 3:case 5:break;default:throw Error(r(329))}if((t&62914560)===t&&(i=ll+300-Mt(),10<i)){if(ra(s,t,Bt,!ia),io(s,0,!0)!==0)break e;jn=t,s.timeoutHandle=Fc(af.bind(null,s,n,Et,cl,Tc,t,Bt,Ya,ks,ia,l,"Throttled",-0,0),i);break e}af(s,n,Et,cl,Tc,t,Bt,Ya,ks,ia,l,null,-0,0)}}break}while(!0);An(e)}function af(e,t,n,s,i,l,d,p,k,R,L,V,O,q){e.timeoutHandle=-1;var P=t.subtreeFlags,se=(l&335544064)===l;if(V=null,(se||P&8192||(P&16785408)===16785408)&&(V={stylesheets:null,count:0,imgCount:0,imgBytes:0,suspenseyImages:[],waitingForImages:!0,waitingForViewTransition:!1,unsuspend:pn},It=null,Fp(t,l,V),se&&(P=V,se=e.containerInfo,se=(se.nodeType===9?se:se.ownerDocument).__reactViewTransition,se!=null&&(P.count++,P.waitingForViewTransition=!0,P=Gi.bind(P),se.finished.then(P,P))),P=(l&62914560)===l?ll-Mt():(l&4194048)===l?$p-Mt():0,P=L1(V,P),P!==null)){jn=l,e.cancelPendingCommit=P(hf.bind(null,e,t,l,n,s,i,d,p,k,R,L,V,null,O,q)),ra(e,l,d,!R);return}hf(e,t,l,n,s,i,d,p,k,R,L,V)}function R0(e){for(var t=e;;){var n=t.tag;if((n===0||n===11||n===15)&&t.flags&16384&&(n=t.updateQueue,n!==null&&(n=n.stores,n!==null)))for(var s=0;s<n.length;s++){var i=n[s],l=i.getSnapshot;i=i.value;try{if(!Ut(l(),i))return!1}catch{return!1}}if(n=t.child,t.subtreeFlags&16384&&n!==null)n.return=t,t=n;else{if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return!0;t=t.return}t.sibling.return=t.return,t=t.sibling}}return!0}function ra(e,t,n,s){t=Pd(e,t),t&=~ol,t&=~Ya,e.suspendedLanes|=t,e.pingedLanes&=~t,s&&(e.warmLanes|=t),s=e.expirationTimes;for(var i=t;0<i;){var l=31-Dt(i),d=1<<l;s[l]=-1,i&=~d}n!==0&&eu(e,n,t)}function ul(){return(Me&6)===0?(Mi(0),!1):!0}function _c(){if(Se!==null){if(De===0)var e=Se.return;else e=Se,zn=Oa=null,Lr(e),ms=null,vi=0,e=Se;for(;e!==null;)Tp(e.alternate,e),e=e.return;Se=null}}function _s(e,t){var n=e.timeoutHandle;return n!==-1&&(e.timeoutHandle=-1,a1(n)),n=e.cancelPendingCommit,n!==null&&(e.cancelPendingCommit=null,n()),jn=0,_c(),Ge=e,Se=n=En(e.current,null),_e=t,De=0,Gt=null,ia=!1,As=$s(e,t),Cc=!1,ks=Bt=ol=Ya=oa=Pe=0,Et=zi=null,Tc=!1,Un=Pd(e,t),xo(),n}function sf(e,t){ke=null,ae.H=Xo,t===fs||t===Eo?(t=uh(),De=3):t===kr?(t=uh(),De=4):De=t===Pr?8:t!==null&&typeof t=="object"&&typeof t.then=="function"?6:1,Gt=t,Se===null&&(Pe=1,Vo(e,Ft(t,e.current)))}function of(){var e=ft.current;return e===null?!0:(_e&4194048)===_e?xt===null:(_e&62914560)===_e||(_e&536870912)!==0?e===xt:!1}function lf(){var e=ae.H;return ae.H=Xo,e===null?Xo:e}function rf(){var e=ae.A;return ae.A=O0,e}function hl(){Pe=4,ia||(_e&4194048)!==_e&&ft.current!==null||(As=!0),(oa&134217727)===0&&(Ya&134217727)===0||Ge===null||ra(Ge,_e,Bt,!1)}function Ec(e,t,n){var s=Me;Me|=2;var i=lf(),l=rf();(Ge!==e||_e!==t)&&(cl=null,_s(e,t)),t=!1;var d=Pe;e:do try{if(De!==0&&Se!==null){var p=Se,k=Gt;switch(De){case 8:_c(),d=6;break e;case 3:case 2:case 9:case 6:ft.current===null&&(t=!0);var R=De;if(De=0,Gt=null,Es(e,p,k,R),n&&As){d=0;break e}break;default:R=De,De=0,Gt=null,Es(e,p,k,R)}}M0(),d=Pe;break}catch(L){sf(e,L)}while(!0);return t&&e.shellSuspendCounter++,zn=Oa=null,Me=s,ae.H=i,ae.A=l,Se===null&&(Ge=null,_e=0,xo()),d}function M0(){for(;Se!==null;)cf(Se)}function q0(e,t){var n=Me;Me|=2;var s=lf(),i=rf();Ge!==e||_e!==t?(cl=null,rl=Mt()+500,_s(e,t)):As=$s(e,t);e:do try{if(De!==0&&Se!==null){t=Se;var l=Gt;t:switch(De){case 1:De=0,Gt=null,Es(e,t,l,1);break;case 2:case 9:if(ch(l)){De=0,Gt=null,df(t);break}t=function(){De!==2&&De!==9||Ge!==e||(De=7),An(e)},l.then(t,t);break e;case 3:De=7;break e;case 4:De=5;break e;case 7:ch(l)?(De=0,Gt=null,df(t)):(De=0,Gt=null,Es(e,t,l,7));break;case 5:var d=null;switch(Se.tag){case 26:d=Se.memoizedState;case 5:case 27:var p=Se;if(d?tm(d):p.stateNode.complete){De=0,Gt=null;var k=p.sibling;if(k!==null)Se=k;else{var R=p.return;R!==null?(Se=R,pl(R)):Se=null}break t}}De=0,Gt=null,Es(e,t,l,5);break;case 6:De=0,Gt=null,Es(e,t,l,6);break;case 8:_c(),Pe=6;break e;default:throw Error(r(462))}}D0();break}catch(L){sf(e,L)}while(!0);return zn=Oa=null,ae.H=s,ae.A=i,Me=n,Se!==null?0:(Ge=null,_e=0,xo(),Pe)}function D0(){for(;Se!==null&&!$v();)cf(Se)}function cf(e){var t=kp(e.alternate,e,Un);e.memoizedProps=e.pendingProps,t===null?pl(e):Se=t}function df(e){var t=e,n=t.alternate;switch(t.tag){case 15:case 0:t=vp(n,t,t.pendingProps,t.type,void 0,_e);break;case 11:t=vp(n,t,t.pendingProps,t.type.render,t.ref,_e);break;case 5:Lr(t);var s=t;s===lt&&(Te?(Co(s),s.tag===5&&s.stateNode!=null&&(Ye=s.stateNode)):(Co(s),Te=!0));default:Tp(n,t),t=Se=Pu(t,Un),t=kp(n,t,Un)}e.memoizedProps=e.pendingProps,t===null?pl(e):Se=t}function Es(e,t,n,s){zn=Oa=null,Lr(t),ms=null,vi=0;var i=t.return;try{if(A0(e,i,t,n,_e)){Pe=1,Vo(e,Ft(n,e.current)),Se=null;return}}catch(l){if(i!==null)throw Se=i,l;Pe=1,Vo(e,Ft(n,e.current)),Se=null;return}t.flags&32768?(Te||s===1?e=!0:As||(_e&536870912)!==0?e=!1:(ia=e=!0,(s===2||s===9||s===3||s===6)&&(s=ft.current,s!==null&&s.tag===13&&(s.flags|=16384))),uf(t,e)):pl(t)}function pl(e){var t=e;do{if((t.flags&32768)!==0){uf(t,ia);return}e=t.return;var n=S0(t.alternate,t,Un);if(n!==null){Se=n;return}if(t=t.sibling,t!==null){Se=t;return}Se=t=e}while(t!==null);Pe===0&&(Pe=5)}function uf(e,t){do{var n=N0(e.alternate,e);if(n!==null){n.flags&=32767,Se=n;return}if(n=e.return,n!==null&&(n.flags|=32768,n.subtreeFlags=0,n.deletions=null),!t&&(e=e.sibling,e!==null)){Se=e;return}Se=e=n}while(e!==null);Pe=6,Se=null}function hf(e,t,n,s,i,l,d,p,k,R,L,V){e.cancelPendingCommit=null;do fl();while(Je!==0);if((Me&6)!==0)throw Error(r(327));if(t!==null){if(t===e.current)throw Error(r(177));e===Ge&&(Se=Ge=null,_e=0),Xa=t,rn=e,jn=n,Nc=i,ef=s,L0(e,t,n,d,p,k,V)}}function L0(e,t,n,s,i,l,d){var p=t.lanes|t.childLanes;if(Sc=p,p|=ur,cy(e,n,p,s,i,l),Ts=null,(n&335544064)===n?(Ss=u0(e),s=10262):(Ss=null,s=10256),(t.subtreeFlags&s)!==0||(t.flags&s)!==0?(e.callbackNode=null,e.callbackPriority=0,Y0(to,function(){return Mc(),null})):(e.callbackNode=null,e.callbackPriority=0),$o=!1,s=(t.flags&13878)!==0,(t.subtreeFlags&13878)!==0||s){s=ae.T,ae.T=null,i=he.p,he.p=2,l=Me,Me|=4;try{_0(e,t,n)}finally{Me=l,he.p=i,ae.T=s}}Je=1,$o?Cs=c1(d,e.containerInfo,Ss,Oc,zc,H0,Rc,Mc,U0):(Oc(),zc(),Rc())}function U0(e){if(Je!==0){var t=rn.onRecoverableError;t(e,{componentStack:null})}}function H0(){Je===3&&(Je=0,Zp(Xa,rn),Je=4)}function Oc(){if(Je===1){Je=0;var e=rn,t=Xa,n=jn,s=(t.flags&13878)!==0;if((t.subtreeFlags&13878)!==0||s){s=ae.T,ae.T=null;var i=he.p;he.p=2;var l=Me;Me|=4;try{_i=nl=!1,Vp(t,e,n),n=Qc;var d=Bu(e.containerInfo),p=n.focusedElem,k=n.selectionRange;if(d!==p&&p&&p.ownerDocument&&Gu(p.ownerDocument.documentElement,p)){if(k!==null&&or(p)){var R=k.start,L=k.end;if(L===void 0&&(L=R),"selectionStart"in p)p.selectionStart=R,p.selectionEnd=Math.min(L,p.value.length);else{var V=p.ownerDocument||document,O=V&&V.defaultView||window;if(O.getSelection){var q=O.getSelection(),P=p.textContent.length,se=Math.min(k.start,P),Ce=k.end===void 0?se:Math.min(k.end,P);!q.extend&&se>Ce&&(d=Ce,Ce=se,se=d);var z=Iu(p,se),T=Iu(p,Ce);if(z&&T&&(q.rangeCount!==1||q.anchorNode!==z.node||q.anchorOffset!==z.offset||q.focusNode!==T.node||q.focusOffset!==T.offset)){var M=V.createRange();M.setStart(z.node,z.offset),q.removeAllRanges(),se>Ce?(q.addRange(M),q.extend(T.node,T.offset)):(M.setEnd(T.node,T.offset),q.addRange(M))}}}}for(V=[],q=p;q=q.parentNode;)q.nodeType===1&&V.push({element:q,left:q.scrollLeft,top:q.scrollTop});for(typeof p.focus=="function"&&p.focus(),p=0;p<V.length;p++){var X=V[p];X.element.scrollLeft=X.left,X.element.scrollTop=X.top}}Us=!!Vc,Qc=Vc=null}finally{Me=l,he.p=i,ae.T=s}}e.current=t,Je=2}}function zc(){if(Je===2){Je=0;var e=rn,t=Xa,n=(t.flags&8772)!==0;if((t.subtreeFlags&8772)!==0||n){n=ae.T,ae.T=null;var s=he.p;he.p=2;var i=Me;Me|=4;try{Hp(e,t.alternate,t)}finally{Me=i,he.p=s,ae.T=n}}Je=3}}function Rc(){if(Je===4||Je===3){Je=0;var e=Cs;Cs=null,ey();var t=rn,n=Xa,s=jn,i=ef,l=(s&335544064)===s?10262:10256;if((n.subtreeFlags&l)!==0||(n.flags&l)!==0?Je=5:(Je=0,Xa=rn=null,pf(t,t.pendingLanes)),l=t.pendingLanes,l===0&&(la=null),Bl(s),n=n.stateNode,qt&&typeof qt.onCommitFiberRoot=="function")try{qt.onCommitFiberRoot(Ps,n,void 0,(n.current.flags&128)===128)}catch{}if(i!==null){n=ae.T,l=he.p,he.p=2,ae.T=null;try{for(var d=t.onRecoverableError,p=0;p<i.length;p++){var k=i[p];d(k.value,{componentStack:k.stack})}}finally{ae.T=n,he.p=l}}if(i=Ts,d=Ss,Ss=null,i!==null&&(Ts=null,d===null&&(d=[]),e!==null))for(k=0;k<i.length;k++)n=(0,i[k])(d),n!==void 0&&e.finished.finally(n);(jn&3)!==0&&fl(),An(t),l=t.pendingLanes,(s&261930)!==0&&(l&42)!==0?t===dl?Ri++:(Ri=0,dl=t):(Ri=0,dl=null),Mi(0)}}function pf(e,t){(e.pooledCacheLanes&=t)===0&&(t=e.pooledCache,t!=null&&(e.pooledCache=null,fi(t)))}function fl(){return Cs!==null&&(Cs.skipTransition(),Cs=null),Oc(),zc(),Rc(),Mc()}function Mc(){if(Je!==5)return!1;var e=rn,t=Sc;Sc=0;var n=Bl(jn),s=ae.T,i=he.p;try{he.p=32>n?32:n,ae.T=null,n=Nc,Nc=null;var l=rn,d=jn;if(Je=0,Xa=rn=null,jn=0,(Me&6)!==0)throw Error(r(331));var p=Me;if(Me|=4,Wp(l.current),Kp(l,l.current,d,n),Me=p,Mi(0,!1),qt&&typeof qt.onPostCommitFiberRoot=="function")try{qt.onPostCommitFiberRoot(Ps,l)}catch{}return!0}finally{he.p=i,ae.T=s,pf(e,t)}}function ff(e,t,n){t=Ft(n,t),t=Wr(e.stateNode,t,2),e=$n(e,t,2),e!==null&&(ei(e,2),An(e))}function Le(e,t,n){if(e.tag===3)ff(e,e,n);else for(;t!==null;){if(t.tag===3){ff(t,e,n);break}else if(t.tag===1){var s=t.stateNode;if(typeof t.type.getDerivedStateFromError=="function"||typeof s.componentDidCatch=="function"&&(la===null||!la.has(s))){e=Ft(n,e),n=cp(2),s=$n(t,n,2),s!==null&&(dp(n,s,t,e),ei(s,2),An(s));break}}t=t.return}}function qc(e,t,n){var s=e.pingCache;if(s===null){s=e.pingCache=new z0;var i=new Set;s.set(t,i)}else i=s.get(t),i===void 0&&(i=new Set,s.set(t,i));i.has(n)||(Cc=!0,i.add(n),e=I0.bind(null,e,t,n),t.then(e,e))}function I0(e,t,n){var s=e.pingCache;s!==null&&s.delete(t),e.pingedLanes|=e.suspendedLanes&n,e.warmLanes&=~n,Ge===e&&(_e&n)===n&&((Pe===4||Pe===3&&(_e&62914560)===_e&&300>Mt()-ll)&&(Me&2)===0?_s(e,0):ol|=n,ks===_e&&(ks=0)),An(e)}function mf(e,t){t===0&&(t=$d()),e=Na(e,t),e!==null&&(ei(e,t),An(e))}function G0(e){var t=e.memoizedState,n=0;t!==null&&(n=t.retryLane),mf(e,n)}function B0(e,t){var n=0;switch(e.tag){case 31:case 13:var s=e.stateNode,i=e.memoizedState;i!==null&&(n=i.retryLane);break;case 19:s=e.stateNode;break;case 22:s=e.stateNode._retryCache;break;default:throw Error(r(314))}s!==null&&s.delete(t),mf(e,n)}function Y0(e,t){return Ul(e,t)}var Os=null,zs=null,Dc=!1,ml=!1,Lc=!1,ca=0;function An(e){e!==zs&&e.next===null&&(zs===null?Os=zs=e:zs=zs.next=e),ml=!0,Dc||(Dc=!0,V0())}function Mi(e,t){if(!Lc&&ml){Lc=!0;do for(var n=!1,s=Os;s!==null;){if(e!==0){var i=s.pendingLanes;if(i===0)var l=0;else{var d=s.suspendedLanes,p=s.pingedLanes;l=(1<<31-Dt(42|e)+1)-1,l&=i&~(d&~p),l=l&201326741?l&201326741|1:l?l|2:0}l!==0&&(n=!0,xf(s,l))}else l=_e,l=io(s,s===Ge?l:0,s.cancelPendingCommit!==null||s.timeoutHandle!==-1),(l&3)===0||$s(s,l)||(n=!0,xf(s,l));s=s.next}while(n);Lc=!1}}function X0(){gf()}function gf(){ml=Dc=!1;var e=0;ca!==0&&n1()&&(e=ca);for(var t=Mt(),n=null,s=Os;s!==null;){var i=s.next,l=vf(s,t);l===0?(s.next=null,n===null?Os=i:n.next=i,i===null&&(zs=n)):(n=s,(e!==0||(l&3)!==0)&&(ml=!0)),s=i}Je!==0&&Je!==5||Mi(e),ca!==0&&(ca=0)}function vf(e,t){for(var n=e.suspendedLanes,s=e.pingedLanes,i=e.expirationTimes,l=e.pendingLanes&-62914561;0<l;){var d=31-Dt(l),p=1<<d,k=i[d];k===-1?((p&n)===0||(p&s)!==0)&&(i[d]=ry(p,t)):k<=t&&(e.expiredLanes|=p),l&=~p}if(t=Ge,n=_e,n=io(e,e===t?n:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),s=e.callbackNode,n===0||e===t&&(De===2||De===9)||e.cancelPendingCommit!==null)return s!==null&&s!==null&&Hl(s),e.callbackNode=null,e.callbackPriority=0;if((n&3)===0||$s(e,n)){if(t=n&-n,t===e.callbackPriority)return t;switch(s!==null&&Hl(s),Bl(n)){case 2:case 8:n=Jd;break;case 32:n=to;break;case 268435456:n=Wd;break;default:n=to}return s=yf.bind(null,e),n=Ul(n,s),e.callbackPriority=t,e.callbackNode=n,t}return s!==null&&s!==null&&Hl(s),e.callbackPriority=2,e.callbackNode=null,2}function yf(e,t){if(Je!==0&&Je!==5)return e.callbackNode=null,e.callbackPriority=0,null;var n=e.callbackNode;if(fl()&&e.callbackNode!==n)return null;var s=_e;return s=io(e,e===Ge?s:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),s===0?null:(nf(e,s,t),vf(e,Mt()),e.callbackNode!=null&&e.callbackNode===n?yf.bind(null,e):null)}function xf(e,t){if(fl())return null;nf(e,t,!0)}function V0(){s1(function(){(Me&6)!==0?Ul(Fd,X0):gf()})}function Uc(){if(ca===0){var e=Ma;e===0&&(e=no,no<<=1,(no&261888)===0&&(no=256)),ca=e}return ca}function wf(e){return e==null||typeof e=="symbol"||typeof e=="boolean"?null:typeof e=="function"?e:uo(e)}function Q0(e,t,n,s,i){if(t==="submit"&&n&&n.stateNode===i){var l=wf((i[Tt]||null).action),d=s.submitter;d&&(t=(t=d[Tt]||null)?wf(t.formAction):d.getAttribute("formAction"),t!==null&&(l=t,d=null));var p=new mo("action","action",null,s,i);e.push({event:p,listeners:[{instance:null,listener:function(){if(s.defaultPrevented){if(ca!==0){var k=new FormData(i,d);Qr(n,{pending:!0,data:k,method:i.method,action:l},null,k)}}else typeof l=="function"&&(p.preventDefault(),k=new FormData(i,d),Qr(n,{pending:!0,data:k,method:i.method,action:l},l,k))},currentTarget:i}]})}}for(var Hc=0;Hc<dr.length;Hc++){var Ic=dr[Hc],Z0=Ic.toLowerCase(),K0=Ic[0].toUpperCase()+Ic.slice(1);an(Z0,"on"+K0)}an(Vu,"onAnimationEnd"),an(Qu,"onAnimationIteration"),an(Zu,"onAnimationStart"),an("dblclick","onDoubleClick"),an("focusin","onFocus"),an("focusout","onBlur"),an(a0,"onTransitionRun"),an(s0,"onTransitionStart"),an(i0,"onTransitionCancel"),an(Ku,"onTransitionEnd"),es("onMouseEnter",["mouseout","mouseover"]),es("onMouseLeave",["mouseout","mouseover"]),es("onPointerEnter",["pointerout","pointerover"]),es("onPointerLeave",["pointerout","pointerover"]),Ca("onChange","change click focusin focusout input keydown keyup selectionchange".split(" ")),Ca("onSelect","focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" ")),Ca("onBeforeInput",["compositionend","keypress","textInput","paste"]),Ca("onCompositionEnd","compositionend focusout keydown keypress keyup mousedown".split(" ")),Ca("onCompositionStart","compositionstart focusout keydown keypress keyup mousedown".split(" ")),Ca("onCompositionUpdate","compositionupdate focusout keydown keypress keyup mousedown".split(" "));var qi="abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "),F0=new Set("beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(qi));function bf(e,t){t=(t&4)!==0;for(var n=0;n<e.length;n++){var s=e[n],i=s.event;s=s.listeners;e:{var l=void 0;if(t)for(var d=s.length-1;0<=d;d--){var p=s[d],k=p.instance,R=p.currentTarget;if(p=p.listener,k!==l&&i.isPropagationStopped())break e;l=p,i.currentTarget=R;try{l(i)}catch(L){yo(L)}i.currentTarget=null,l=k}else for(d=0;d<s.length;d++){if(p=s[d],k=p.instance,R=p.currentTarget,p=p.listener,k!==l&&i.isPropagationStopped())break e;l=p,i.currentTarget=R;try{l(i)}catch(L){yo(L)}i.currentTarget=null,l=k}}}}function Ne(e,t){var n=t[iu];n===void 0&&(n=t[iu]=new Set);var s=e+"__bubble";n.has(s)||(jf(t,e,2,!1),n.add(s))}function Gc(e,t,n){var s=0;t&&(s|=4),jf(n,e,s,t)}var gl="_reactListening"+Math.random().toString(36).slice(2);function Bc(e){if(!e[gl]){e[gl]=!0,ru.forEach(function(n){n!=="selectionchange"&&(F0.has(n)||Gc(n,!1,e),Gc(n,!0,e))});var t=e.nodeType===9?e:e.ownerDocument;t===null||t[gl]||(t[gl]=!0,Gc("selectionchange",!1,t))}}function jf(e,t,n,s){switch(um(t)){case 2:var i=G1;break;case 8:i=B1;break;default:i=rd}n=i.bind(null,t,n,e),i=void 0,!Jl||t!=="touchstart"&&t!=="touchmove"&&t!=="wheel"||(i=!0),s?i!==void 0?e.addEventListener(t,n,{capture:!0,passive:i}):e.addEventListener(t,n,!0):i!==void 0?e.addEventListener(t,n,{passive:i}):e.addEventListener(t,n,!1)}function Yc(e,t,n,s,i){var l=s;if((t&1)===0&&(t&2)===0&&s!==null)e:for(;;){if(s===null)return;var d=s.tag;if(d===3||d===4){var p=s.stateNode.containerInfo;if(p===i)break;if(d===4)for(d=s.return;d!==null;){var k=d.tag;if((k===3||k===4)&&d.stateNode.containerInfo===i)return;d=d.return}for(;p!==null;){if(d=ka(p),d===null)return;if(k=d.tag,k===5||k===6||k===26||k===27){s=l=d;continue e}p=p.parentNode}}s=s.return}bu(function(){var R=l,L=Kl(n),V=[];e:{var O=Fu.get(e);if(O!==void 0){var q=mo,P=e;switch(e){case"keypress":if(po(n)===0)break e;case"keydown":case"keyup":q=Ry;break;case"focusin":P="focus",q=er;break;case"focusout":P="blur",q=er;break;case"beforeblur":case"afterblur":q=er;break;case"click":if(n.button===2)break e;case"auxclick":case"dblclick":case"mousedown":case"mousemove":case"mouseup":case"mouseout":case"mouseover":case"contextmenu":q=ku;break;case"drag":case"dragend":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"dragstart":case"drop":q=by;break;case"touchcancel":case"touchend":case"touchmove":case"touchstart":q=Uy;break;case Vu:case Qu:case Zu:q=ky;break;case Ku:q=Iy;break;case"scroll":case"scrollend":q=xy;break;case"wheel":q=By;break;case"copy":case"cut":case"paste":q=Ty;break;case"gotpointercapture":case"lostpointercapture":case"pointercancel":case"pointerdown":case"pointermove":case"pointerout":case"pointerover":case"pointerup":q=Tu;break;case"submit":q=Dy;break;case"toggle":case"beforetoggle":q=Xy}var se=(t&4)!==0,Ce=!se&&(e==="scroll"||e==="scrollend"),z=se?O!==null?O+"Capture":null:O;se=[];for(var T=R,M;T!==null;){var X=T;if(M=X.stateNode,X=X.tag,X!==5&&X!==26&&X!==27||M===null||z===null||(X=ai(T,z),X!=null&&se.push(Di(T,X,M))),Ce)break;T=T.return}0<se.length&&(O=new q(O,P,null,n,L),V.push({event:O,listeners:se}))}}if((t&7)===0){e:{if(q=e==="mouseover"||e==="pointerover",O=e==="mouseout"||e==="pointerout",q&&n!==Zl&&(P=n.relatedTarget||n.fromElement)&&(ka(P)||P[Wa]))break e;(O||q)&&(P=L.window===L?L:(q=L.ownerDocument)?q.defaultView||q.parentWindow:window,O?(q=n.relatedTarget||n.toElement,O=R,q=q?ka(q):null,q!==null&&(Ce=m(q),se=q.tag,q!==Ce||se!==5&&se!==27&&se!==6)&&(q=null)):(O=null,q=R),O!==q&&(se=ku,X="onMouseLeave",z="onMouseEnter",T="mouse",(e==="pointerout"||e==="pointerover")&&(se=Tu,X="onPointerLeave",z="onPointerEnter",T="pointer"),Ce=O==null?P:ni(O),M=q==null?P:ni(q),P=new se(X,T+"leave",O,n,L),P.target=Ce,P.relatedTarget=M,X=null,ka(L)===R&&(se=new se(z,T+"enter",q,n,L),se.target=M,se.relatedTarget=Ce,X=se),Ce=X,se=O&&q?S(O,q,J0):null,O!==null&&Af(V,P,O,se,!1),q!==null&&Ce!==null&&Af(V,Ce,q,se,!0)))}e:{if(O=R?ni(R):window,q=O.nodeName&&O.nodeName.toLowerCase(),q==="select"||q==="input"&&O.type==="file")var te=Mu;else if(zu(O))if(qu)te=e0;else{te=Py;var Ee=Wy}else q=O.nodeName,!q||q.toLowerCase()!=="input"||O.type!=="checkbox"&&O.type!=="radio"?R&&Ql(R.elementType)&&(te=Mu):te=$y;if(te&&(te=te(e,R))){Ru(V,te,n,L);break e}Ee&&Ee(e,O,R)}switch(Ee=R?ni(R):window,e){case"focusin":(zu(Ee)||Ee.contentEditable==="true")&&(os=Ee,lr=R,ui=null);break;case"focusout":ui=lr=os=null;break;case"mousedown":rr=!0;break;case"contextmenu":case"mouseup":case"dragend":rr=!1,Yu(V,n,L);break;case"selectionchange":if(n0)break;case"keydown":case"keyup":Yu(V,n,L)}var ue;if(nr)e:{switch(e){case"compositionstart":var pe="onCompositionStart";break e;case"compositionend":pe="onCompositionEnd";break e;case"compositionupdate":pe="onCompositionUpdate";break e}pe=void 0}else is?Eu(e,n)&&(pe="onCompositionEnd"):e==="keydown"&&n.keyCode===229&&(pe="onCompositionStart");pe&&(Su&&n.locale!=="ko"&&(is||pe!=="onCompositionStart"?pe==="onCompositionEnd"&&is&&(ue=ju()):(Xn=L,Wl="value"in Xn?Xn.value:Xn.textContent,is=!0)),Ee=vl(R,pe),0<Ee.length&&(pe=new Cu(pe,e,null,n,L),V.push({event:pe,listeners:Ee}),ue?pe.data=ue:(ue=Ou(n),ue!==null&&(pe.data=ue)))),(ue=Qy?Zy(e,n):Ky(e,n))&&(pe=vl(R,"onBeforeInput"),0<pe.length&&(Ee=new Cu("onBeforeInput","beforeinput",null,n,L),V.push({event:Ee,listeners:pe}),Ee.data=ue)),Q0(V,e,R,n,L)}bf(V,t)})}function Di(e,t,n){return{instance:e,listener:t,currentTarget:n}}function vl(e,t){for(var n=t+"Capture",s=[];e!==null;){var i=e,l=i.stateNode;if(i=i.tag,i!==5&&i!==26&&i!==27||l===null||(i=ai(e,n),i!=null&&s.unshift(Di(e,i,l)),i=ai(e,t),i!=null&&s.push(Di(e,i,l))),e.tag===3)return s;e=e.return}return[]}function J0(e){if(e===null)return null;do e=e.return;while(e&&e.tag!==5&&e.tag!==27);return e||null}function Af(e,t,n,s,i){for(var l=t._reactName,d=[];n!==null&&n!==s;){var p=n,k=p.alternate,R=p.stateNode;if(p=p.tag,k!==null&&k===s)break;p!==5&&p!==26&&p!==27||R===null||(k=R,i?(R=ai(n,l),R!=null&&d.unshift(Di(n,R,k))):i||(R=ai(n,l),R!=null&&d.push(Di(n,R,k)))),n=n.return}d.length!==0&&e.push({event:t,listeners:d})}var W0=/\r\n?/g,P0=/\u0000|\uFFFD/g;function kf(e){return(typeof e=="string"?e:""+e).replace(W0,`
`).replace(P0,"")}function Cf(e,t){return t=kf(t),kf(e)===t}function Ue(e,t,n,s,i,l){switch(n){case"children":if(typeof s=="string")t==="body"||t==="textarea"&&s===""||ns(e,s);else if(typeof s=="number"||typeof s=="bigint")t!=="body"&&ns(e,""+s);else return;break;case"className":co(e,"class",s);break;case"tabIndex":co(e,"tabindex",s);break;case"dir":case"role":case"viewBox":case"width":case"height":co(e,n,s);break;case"style":xu(e,s,l);return;case"data":if(t!=="object"){co(e,"data",s);break}case"src":case"href":if(s===""&&(t!=="a"||n!=="href")){e.removeAttribute(n);break}if(s==null||typeof s=="function"||typeof s=="symbol"||typeof s=="boolean"){e.removeAttribute(n);break}s=uo(s),e.setAttribute(n,s);break;case"action":case"formAction":if(typeof s=="function"){e.setAttribute(n,"javascript:throw new Error('A React form was unexpectedly submitted. If you called form.submit() manually, consider using form.requestSubmit() instead. If you\\'re trying to use event.stopPropagation() in a submit event handler, consider also calling event.preventDefault().')");break}else typeof l=="function"&&(n==="formAction"?(t!=="input"&&Ue(e,t,"name",i.name,i,null),Ue(e,t,"formEncType",i.formEncType,i,null),Ue(e,t,"formMethod",i.formMethod,i,null),Ue(e,t,"formTarget",i.formTarget,i,null)):(Ue(e,t,"encType",i.encType,i,null),Ue(e,t,"method",i.method,i,null),Ue(e,t,"target",i.target,i,null)));if(s==null||typeof s=="symbol"||typeof s=="boolean"){e.removeAttribute(n);break}s=uo(s),e.setAttribute(n,s);break;case"onClick":s!=null&&(e.onclick=pn);return;case"onScroll":s!=null&&Ne("scroll",e);return;case"onScrollEnd":s!=null&&Ne("scrollend",e);return;case"dangerouslySetInnerHTML":if(s!=null){if(typeof s!="object"||!("__html"in s))throw Error(r(61));if(n=s.__html,n!=null){if(i.children!=null)throw Error(r(60));l?.__html!==n&&(e.innerHTML=n)}}break;case"multiple":e.multiple=s&&typeof s!="function"&&typeof s!="symbol";break;case"muted":e.muted=s&&typeof s!="function"&&typeof s!="symbol";break;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"defaultValue":case"defaultChecked":case"innerHTML":case"ref":break;case"autoFocus":break;case"xlinkHref":if(s==null||typeof s=="function"||typeof s=="boolean"||typeof s=="symbol"){e.removeAttribute("xlink:href");break}n=uo(s),e.setAttributeNS("http://www.w3.org/1999/xlink","xlink:href",n);break;case"contentEditable":case"spellCheck":case"draggable":case"value":case"autoReverse":case"externalResourcesRequired":case"focusable":case"preserveAlpha":s!=null&&typeof s!="function"&&typeof s!="symbol"?e.setAttribute(n,s):e.removeAttribute(n);break;case"inert":case"allowFullScreen":case"async":case"autoPlay":case"controls":case"credentialless":case"default":case"defer":case"disabled":case"disablePictureInPicture":case"disableRemotePlayback":case"formNoValidate":case"hidden":case"loop":case"noModule":case"noValidate":case"open":case"playsInline":case"readOnly":case"required":case"reversed":case"scoped":case"seamless":case"itemScope":s&&typeof s!="function"&&typeof s!="symbol"?e.setAttribute(n,""):e.removeAttribute(n);break;case"capture":case"download":s===!0?e.setAttribute(n,""):s!==!1&&s!=null&&typeof s!="function"&&typeof s!="symbol"?e.setAttribute(n,s):e.removeAttribute(n);break;case"cols":case"rows":case"size":case"span":s!=null&&typeof s!="function"&&typeof s!="symbol"&&!isNaN(s)&&1<=s?e.setAttribute(n,s):e.removeAttribute(n);break;case"rowSpan":case"start":s==null||typeof s=="function"||typeof s=="symbol"||isNaN(s)?e.removeAttribute(n):e.setAttribute(n,s);break;case"popover":Ne("beforetoggle",e),Ne("toggle",e),ro(e,"popover",s);break;case"xlinkActuate":Tn(e,"http://www.w3.org/1999/xlink","xlink:actuate",s);break;case"xlinkArcrole":Tn(e,"http://www.w3.org/1999/xlink","xlink:arcrole",s);break;case"xlinkRole":Tn(e,"http://www.w3.org/1999/xlink","xlink:role",s);break;case"xlinkShow":Tn(e,"http://www.w3.org/1999/xlink","xlink:show",s);break;case"xlinkTitle":Tn(e,"http://www.w3.org/1999/xlink","xlink:title",s);break;case"xlinkType":Tn(e,"http://www.w3.org/1999/xlink","xlink:type",s);break;case"xmlBase":Tn(e,"http://www.w3.org/XML/1998/namespace","xml:base",s);break;case"xmlLang":Tn(e,"http://www.w3.org/XML/1998/namespace","xml:lang",s);break;case"xmlSpace":Tn(e,"http://www.w3.org/XML/1998/namespace","xml:space",s);break;case"is":ro(e,"is",s);break;case"innerText":case"textContent":return;default:if(!(2<n.length)||n[0]!=="o"&&n[0]!=="O"||n[1]!=="n"&&n[1]!=="N")n=vy.get(n)||n,ro(e,n,s);else return}ze=!0}function Xc(e,t,n,s,i,l){switch(n){case"style":xu(e,s,l);return;case"dangerouslySetInnerHTML":if(s!=null){if(typeof s!="object"||!("__html"in s))throw Error(r(61));if(n=s.__html,n!=null){if(i.children!=null)throw Error(r(60));l?.__html!==n&&(e.innerHTML=n)}}break;case"children":if(typeof s=="string")ns(e,s);else if(typeof s=="number"||typeof s=="bigint")ns(e,""+s);else return;break;case"onScroll":s!=null&&Ne("scroll",e);return;case"onScrollEnd":s!=null&&Ne("scrollend",e);return;case"onClick":s!=null&&(e.onclick=pn);return;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"innerHTML":case"ref":return;case"innerText":case"textContent":return;default:if(!cu.hasOwnProperty(n))e:{if(n[0]==="o"&&n[1]==="n"&&(i=n.endsWith("Capture"),l=n.slice(2,i?n.length-7:void 0),t=e[Tt]||null,t=t!=null?t[n]:null,typeof t=="function"&&e.removeEventListener(l,t,i),typeof s=="function")){typeof t!="function"&&t!==null&&(n in e?e[n]=null:e.hasAttribute(n)&&e.removeAttribute(n)),e.addEventListener(l,s,i);break e}ze=!0,n in e?e[n]=s:s===!0?e.setAttribute(n,""):ro(e,n,s)}return}ze=!0}function vt(e,t,n){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"img":Ne("error",e),Ne("load",e);var s=!1,i=!1,l;for(l in n)if(n.hasOwnProperty(l)){var d=n[l];if(d!=null)switch(l){case"src":s=!0;break;case"srcSet":i=!0;break;case"children":case"dangerouslySetInnerHTML":throw Error(r(137,t));default:Ue(e,t,l,d,n,null)}}i&&Ue(e,t,"srcSet",n.srcSet,n,null),s&&Ue(e,t,"src",n.src,n,null);return;case"input":Ne("invalid",e);var p=l=d=i=null,k=null,R=null;for(s in n)if(n.hasOwnProperty(s)){var L=n[s];if(L!=null)switch(s){case"name":i=L;break;case"type":d=L;break;case"checked":k=L;break;case"defaultChecked":R=L;break;case"value":l=L;break;case"defaultValue":p=L;break;case"children":case"dangerouslySetInnerHTML":if(L!=null)throw Error(r(137,t));break;default:Ue(e,t,s,L,n,null)}}mu(e,l,p,k,R,d,i,!1);return;case"select":Ne("invalid",e),s=d=l=null;for(i in n)if(n.hasOwnProperty(i)&&(p=n[i],p!=null))switch(i){case"value":l=p;break;case"defaultValue":d=p;break;case"multiple":s=p;default:Ue(e,t,i,p,n,null)}t=l,n=d,e.multiple=!!s,t!=null?ts(e,!!s,t,!1):n!=null&&ts(e,!!s,n,!0);return;case"textarea":Ne("invalid",e),l=i=s=null;for(d in n)if(n.hasOwnProperty(d)&&(p=n[d],p!=null))switch(d){case"value":s=p;break;case"defaultValue":i=p;break;case"children":l=p;break;case"dangerouslySetInnerHTML":if(p!=null)throw Error(r(91));break;default:Ue(e,t,d,p,n,null)}vu(e,s,i,l);return;case"option":for(k in n)n.hasOwnProperty(k)&&(s=n[k],s!=null)&&(k==="selected"?e.selected=s&&typeof s!="function"&&typeof s!="symbol":Ue(e,t,k,s,n,null));return;case"dialog":Ne("beforetoggle",e),Ne("toggle",e),Ne("cancel",e),Ne("close",e);break;case"iframe":case"object":Ne("load",e);break;case"video":case"audio":for(s=0;s<qi.length;s++)Ne(qi[s],e);break;case"image":Ne("error",e),Ne("load",e);break;case"details":Ne("toggle",e);break;case"embed":case"source":case"link":Ne("error",e),Ne("load",e);case"area":case"base":case"br":case"col":case"hr":case"keygen":case"meta":case"param":case"track":case"wbr":case"menuitem":for(R in n)if(n.hasOwnProperty(R)&&(s=n[R],s!=null))switch(R){case"children":case"dangerouslySetInnerHTML":throw Error(r(137,t));default:Ue(e,t,R,s,n,null)}return;default:if(Ql(t)){for(L in n)n.hasOwnProperty(L)&&(s=n[L],s!==void 0&&Xc(e,t,L,s,n,void 0));return}}for(p in n)n.hasOwnProperty(p)&&(s=n[p],s!=null&&Ue(e,t,p,s,n,null))}var $0={};function e1(e,t,n,s){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"input":var i=null,l=null,d=null,p=null,k=null,R=null,L=null;for(q in n){var V=n[q];if(n.hasOwnProperty(q)&&V!=null)switch(q){case"checked":break;case"value":break;case"defaultValue":k=V;default:s.hasOwnProperty(q)||Ue(e,t,q,null,s,V)}}for(var O in s){var q=s[O];if(V=n[O],s.hasOwnProperty(O)&&(q!=null||V!=null))switch(O){case"type":q!==V&&(ze=!0),l=q;break;case"name":q!==V&&(ze=!0),i=q;break;case"checked":q!==V&&(ze=!0),R=q;break;case"defaultChecked":q!==V&&(ze=!0),L=q;break;case"value":q!==V&&(ze=!0),d=q;break;case"defaultValue":q!==V&&(ze=!0),p=q;break;case"children":case"dangerouslySetInnerHTML":if(q!=null)throw Error(r(137,t));break;default:q!==V&&Ue(e,t,O,q,s,V)}}Xl(e,d,p,k,R,L,l,i);return;case"select":q=d=p=O=null;for(l in n)if(k=n[l],n.hasOwnProperty(l)&&k!=null)switch(l){case"value":break;case"multiple":q=k;default:s.hasOwnProperty(l)||Ue(e,t,l,null,s,k)}for(i in s)if(l=s[i],k=n[i],s.hasOwnProperty(i)&&(l!=null||k!=null))switch(i){case"value":l!==k&&(ze=!0),O=l;break;case"defaultValue":l!==k&&(ze=!0),p=l;break;case"multiple":l!==k&&(ze=!0),d=l;default:l!==k&&Ue(e,t,i,l,s,k)}t=p,n=d,s=q,O!=null?ts(e,!!n,O,!1):!!s!=!!n&&(t!=null?ts(e,!!n,t,!0):ts(e,!!n,n?[]:"",!1));return;case"textarea":q=O=null;for(p in n)if(i=n[p],n.hasOwnProperty(p)&&i!=null&&!s.hasOwnProperty(p))switch(p){case"value":break;case"children":break;default:Ue(e,t,p,null,s,i)}for(d in s)if(i=s[d],l=n[d],s.hasOwnProperty(d)&&(i!=null||l!=null))switch(d){case"value":i!==l&&(ze=!0),O=i;break;case"defaultValue":i!==l&&(ze=!0),q=i;break;case"children":break;case"dangerouslySetInnerHTML":if(i!=null)throw Error(r(91));break;default:i!==l&&Ue(e,t,d,i,s,l)}gu(e,O,q);return;case"option":for(var P in n)O=n[P],n.hasOwnProperty(P)&&O!=null&&!s.hasOwnProperty(P)&&(P==="selected"?e.selected=!1:Ue(e,t,P,null,s,O));for(k in s)O=s[k],q=n[k],s.hasOwnProperty(k)&&O!==q&&(O!=null||q!=null)&&(k==="selected"?(O!==q&&(ze=!0),e.selected=O&&typeof O!="function"&&typeof O!="symbol"):Ue(e,t,k,O,s,q));return;case"img":case"link":case"area":case"base":case"br":case"col":case"embed":case"hr":case"keygen":case"meta":case"param":case"source":case"track":case"wbr":case"menuitem":for(var se in n)O=n[se],n.hasOwnProperty(se)&&O!=null&&!s.hasOwnProperty(se)&&Ue(e,t,se,null,s,O);for(R in s)if(O=s[R],q=n[R],s.hasOwnProperty(R)&&O!==q&&(O!=null||q!=null))switch(R){case"children":case"dangerouslySetInnerHTML":if(O!=null)throw Error(r(137,t));break;default:Ue(e,t,R,O,s,q)}return;default:if(Ql(t)){for(var Ce in n)O=n[Ce],n.hasOwnProperty(Ce)&&O!==void 0&&!s.hasOwnProperty(Ce)&&Xc(e,t,Ce,void 0,s,O);for(L in s)O=s[L],q=n[L],!s.hasOwnProperty(L)||O===q||O===void 0&&q===void 0||Xc(e,t,L,O,s,q);return}}for(var z in n)O=n[z],n.hasOwnProperty(z)&&O!=null&&!s.hasOwnProperty(z)&&Ue(e,t,z,null,s,O);for(V in s)O=s[V],q=n[V],!s.hasOwnProperty(V)||O===q||O==null&&q==null||Ue(e,t,V,O,s,q)}function Tf(e){switch(e){case"css":case"script":case"font":case"img":case"image":case"input":case"link":return!0;default:return!1}}function t1(){if(typeof performance.getEntriesByType=="function"){for(var e=0,t=0,n=performance.getEntriesByType("resource"),s=0;s<n.length;s++){var i=n[s],l=i.transferSize,d=i.initiatorType,p=i.duration;if(l&&p&&Tf(d)){for(d=0,p=i.responseEnd,s+=1;s<n.length;s++){var k=n[s],R=k.startTime;if(R>p)break;var L=k.transferSize,V=k.initiatorType;L&&Tf(V)&&(k=k.responseEnd,d+=L*(k<p?1:(p-R)/(k-R)))}if(--s,t+=8*(l+d)/(i.duration/1e3),e++,10<e)break}}if(0<e)return t/e/1e6}return navigator.connection&&(e=navigator.connection.downlink,typeof e=="number")?e:5}var Vc=null,Qc=null;function Li(e){return e.nodeType===9?e:e.ownerDocument}function Sf(e){switch(e){case"http://www.w3.org/2000/svg":return 1;case"http://www.w3.org/1998/Math/MathML":return 2;default:return 0}}function Nf(e,t){if(e===0)switch(t){case"svg":return 1;case"math":return 2;default:return 0}return e===1&&t==="foreignObject"?0:e}function _f(e,t,n,s){return n=Li(n).createElement(e),n[ht]=s,n[Tt]=t,vt(n,e,t),ot(n),n}function Zc(e,t){return e==="textarea"||e==="noscript"||typeof t.children=="string"||typeof t.children=="number"||typeof t.children=="bigint"||typeof t.dangerouslySetInnerHTML=="object"&&t.dangerouslySetInnerHTML!==null&&t.dangerouslySetInnerHTML.__html!=null}var Kc=null;function n1(){var e=window.event;return e&&e.type==="popstate"?e===Kc?!1:(Kc=e,!0):(Kc=null,!1)}var Fc=typeof setTimeout=="function"?setTimeout:void 0,a1=typeof clearTimeout=="function"?clearTimeout:void 0,Ef=typeof Promise=="function"?Promise:void 0,Of=typeof requestAnimationFrame=="function"?requestAnimationFrame:Fc,s1=typeof queueMicrotask=="function"?queueMicrotask:typeof Ef<"u"?function(e){return Ef.resolve(null).then(e).catch(i1)}:Fc;function i1(e){setTimeout(function(){throw e})}function da(e){return e==="head"}function zf(e,t){var n=t,s=0;do{var i=n.nextSibling;if(e.removeChild(n),i&&i.nodeType===8)if(n=i.data,n==="/$"||n==="/&"){if(s===0){e.removeChild(i),Hs(t);return}s--}else if(n==="$"||n==="$?"||n==="$~"||n==="$!"||n==="&")s++;else if(n==="html")ad(e.ownerDocument.documentElement);else if(n==="head"){n=e.ownerDocument.head,ad(n);for(var l=n.firstChild;l;){var d=l.nextSibling,p=l.nodeName;l[ti]||p==="SCRIPT"||p==="STYLE"||p==="LINK"&&l.rel.toLowerCase()==="stylesheet"||n.removeChild(l),l=d}}else n==="body"&&ad(e.ownerDocument.body);n=i}while(n);Hs(t)}function Rf(e,t){var n=e;e=0;do{var s=n.nextSibling;if(n.nodeType===1?t?(n._stashedDisplay=n.style.display,n.style.display="none"):(n.style.display=n._stashedDisplay||"",n.getAttribute("style")===""&&n.removeAttribute("style")):n.nodeType===3&&(t?(n._stashedText=n.nodeValue,n.nodeValue=""):n.nodeValue=n._stashedText||""),s&&s.nodeType===8)if(n=s.data,n==="/$"){if(e===0)break;e--}else n!=="$"&&n!=="$?"&&n!=="$~"&&n!=="$!"||e++;n=s}while(n)}function Mf(e,t,n){if(t=CSS.escape(t)!==t?"r-"+btoa(t).replace(/=/g,""):t,e.style.viewTransitionName=t,n!=null&&(e.style.viewTransitionClass=n),n=getComputedStyle(e),n.display==="inline"){if(t=e.getClientRects(),t.length===1)var s=1;else for(var i=s=0;i<t.length;i++){var l=t[i];0<l.width&&0<l.height&&s++}s===1&&(e=e.style,e.display=t.length===1?"inline-block":"block",e.marginTop="-"+n.paddingTop,e.marginBottom="-"+n.paddingBottom)}}function qf(e,t){e=e.style,t=t.style;var n=t!=null?t.hasOwnProperty("viewTransitionName")?t.viewTransitionName:t.hasOwnProperty("view-transition-name")?t["view-transition-name"]:null:null;e.viewTransitionName=n==null||typeof n=="boolean"?"":(""+n).trim(),n=t!=null?t.hasOwnProperty("viewTransitionClass")?t.viewTransitionClass:t.hasOwnProperty("view-transition-class")?t["view-transition-class"]:null:null,e.viewTransitionClass=n==null||typeof n=="boolean"?"":(""+n).trim(),e.display==="inline-block"&&(t==null?e.display=e.margin="":(n=t.display,e.display=n==null||typeof n=="boolean"?"":n,n=t.margin,n!=null?e.margin=n:(n=t.hasOwnProperty("marginTop")?t.marginTop:t["margin-top"],e.marginTop=n==null||typeof n=="boolean"?"":n,t=t.hasOwnProperty("marginBottom")?t.marginBottom:t["margin-bottom"],e.marginBottom=t==null||typeof t=="boolean"?"":t)))}function o1(e,t,n){return n=n.ownerDocument.defaultView,{rect:e,abs:t.position==="absolute"||t.position==="fixed",clip:t.clipPath!=="none"||t.overflow!=="visible"||t.filter!=="none"||t.mask!=="none"||t.mask!=="none"||t.borderRadius!=="0px",view:0<=e.bottom&&0<=e.right&&e.top<=n.innerHeight&&e.left<=n.innerWidth}}function Jc(e){var t=e.getBoundingClientRect(),n=getComputedStyle(e);return o1(t,n,e)}function l1(e){return e.documentElement.clientHeight}function r1(e){this.addEventListener("load",e),this.addEventListener("error",e)}function c1(e,t,n,s,i,l,d,p,k){var R=t.nodeType===9?t:t.ownerDocument;try{var L=R.startViewTransition({update:function(){var O=R.defaultView,q=O.navigation&&O.navigation.transition,P=R.fonts.status;s();var se=[];if(P==="loaded"&&(l1(R),R.fonts.status==="loading"&&se.push(R.fonts.ready)),P=se.length,e!==null)for(var Ce=e.suspenseyImages,z=0,T=0;T<Ce.length;T++){var M=Ce[T];if(!M.complete){var X=M.getBoundingClientRect();if(0<X.bottom&&0<X.right&&X.top<O.innerHeight&&X.left<O.innerWidth){if(z+=nm(M),z>wl){se.length=P;break}M=new Promise(r1.bind(M)),se.push(M)}}}if(0<se.length)return O=Promise.race([Promise.all(se),new Promise(function(te){return setTimeout(te,500)})]).then(i,i),(q?Promise.allSettled([q.finished,O]):O).then(l,l);if(i(),q)return q.finished.then(l,l);l()},types:n});R.__reactViewTransition=L;var V=[];return L.ready.then(function(){for(var O=R.documentElement.getAnimations({subtree:!0}),q=0;q<O.length;q++){var P=O[q],se=P.effect,Ce=se.pseudoElement;if(Ce!=null&&Ce.startsWith("::view-transition")){V.push(P),P=se.getKeyframes();for(var z=Ce=void 0,T=!0,M=0;M<P.length;M++){var X=P[M],te=X.width;if(Ce===void 0)Ce=te;else if(Ce!==te){T=!1;break}if(te=X.height,z===void 0)z=te;else if(z!==te){T=!1;break}delete X.width,delete X.height,X.transform==="none"&&delete X.transform}T&&Ce!==void 0&&z!==void 0&&(se.setKeyframes(P),T=getComputedStyle(se.target,se.pseudoElement),T.width!==Ce||T.height!==z)&&(T=P[0],T.width=Ce,T.height=z,T=P[P.length-1],T.width=Ce,T.height=z,se.setKeyframes(P))}}d()},function(O){R.__reactViewTransition===L&&(R.__reactViewTransition=null);try{typeof O=="object"&&O!==null&&O.name==="InvalidStateError"&&(O.message==="View transition was skipped because document visibility state is hidden."||O.message==="Skipping view transition because document visibility state has become hidden."||O.message==="Skipping view transition because viewport size changed."||O.message==="Transition was aborted because of invalid state")&&(O=null),O!==null&&k(O)}finally{s(),i(),d()}}),L.finished.finally(function(){for(var O=0;O<V.length;O++)V[O].cancel();R.__reactViewTransition===L&&(R.__reactViewTransition=null),p()}),L}catch{return s(),i(),d(),null}}function Va(e,t){this._scope=document.documentElement,this._selector="::view-transition-"+e+"("+t+")"}Va.prototype.animate=function(e,t){return t=typeof t=="number"?{duration:t}:U({},t),t.pseudoElement=this._selector,this._scope.animate(e,t)},Va.prototype.getAnimations=function(){for(var e=this._scope,t=this._selector,n=e.getAnimations({subtree:!0}),s=[],i=0;i<n.length;i++){var l=n[i].effect;l!==null&&l.target===e&&l.pseudoElement===t&&s.push(n[i])}return s},Va.prototype.getComputedStyle=function(){return getComputedStyle(this._scope,this._selector)};function Df(e){return{name:e,group:new Va("group",e),imagePair:new Va("image-pair",e),old:new Va("old",e),new:new Va("new",e)}}function Xt(e){this._fragmentFiber=e,this._observers=this._eventListeners=null}Xt.prototype.addEventListener=function(e,t,n){var s=null,i=null;if(!(n!=null&&typeof n!="boolean"&&(s=n.signal||null,s!==null&&s.aborted))){this._eventListeners===null&&(this._eventListeners=[]);var l=this._eventListeners;if(Uf(l,e,t,n)===-1){var d=this,p=t;n!=null&&typeof n!="boolean"&&n.once===!0&&(p=function(k){d.removeEventListener(e,t,n),typeof t=="function"?t.call(this,k):t.handleEvent(k)}),s!==null&&(i=d.removeEventListener.bind(d,e,t,n),s.addEventListener("abort",i,{once:!0}),i=s.removeEventListener.bind(s,"abort",i)),s=Rs(n),l.push({type:e,listener:t,optionsOrUseCapture:n,attachedListener:p,cleanup:i}),y(this._fragmentFiber.child,!1,d1,e,p,s)}this._eventListeners=l}};function d1(e,t,n,s){return N(e).addEventListener(t,n,s),!1}Xt.prototype.removeEventListener=function(e,t,n){var s=this._eventListeners;if(s!==null&&(t=Uf(s,e,t,n),t!==-1)){var i=s[t];n=i.attachedListener;var l=i.cleanup;i=Rs(i.optionsOrUseCapture),y(this._fragmentFiber.child,!1,u1,e,n,i),s.splice(t,1),l!==null&&l()}};function u1(e,t,n,s){return N(e).removeEventListener(t,n,s),!1}function Rs(e){return e!=null&&typeof e!="boolean"&&(e.once===!0||e.signal instanceof AbortSignal)?{capture:e.capture,passive:e.passive}:e}function Lf(e){return e==null?"c=0":typeof e=="boolean"?"c="+(e?"1":"0"):"c="+(e.capture?"1":"0")}function Uf(e,t,n,s){if(e.length===0)return-1;s=Lf(s);for(var i=0;i<e.length;i++){var l=e[i];if(l.type===t&&l.listener===n&&Lf(l.optionsOrUseCapture)===s)return i}return-1}Xt.prototype.dispatchEvent=function(e){var t=w(this._fragmentFiber);if(t===null)return!0;t=N(t);var n=this._eventListeners;if(n!==null&&0<n.length||!e.bubbles){var s=t.nodeType===9?t.createComment(""):document.createTextNode("");if(n)for(var i=0;i<n.length;i++){var l=n[i];s.addEventListener(l.type,l.attachedListener,Rs(l.optionsOrUseCapture))}if(t.appendChild(s),e=s.dispatchEvent(e),n)for(i=0;i<n.length;i++)l=n[i],s.removeEventListener(l.type,l.attachedListener,Rs(l.optionsOrUseCapture));return t.removeChild(s),e}return t.dispatchEvent(e)},Xt.prototype.focus=function(e){y(this._fragmentFiber.child,!0,Hf,e,void 0,void 0)};function Hf(e,t){return e.tag===6?!1:(e=N(e),A1(e,t))}Xt.prototype.focusLast=function(e){var t=[];y(this._fragmentFiber.child,!0,Wc,t,void 0,void 0);for(var n=t.length-1;0<=n&&!Hf(t[n],e);n--);};function Wc(e,t){return t.push(e),!1}Xt.prototype.blur=function(){var e=w(this._fragmentFiber);e!==null&&(e=N(e),e=Li(e).activeElement,e!==null&&y(this._fragmentFiber.child,!1,h1,e,void 0,void 0))};function h1(e,t){return e.tag===6?!1:(e=N(e),e===t||e.contains(t)?(t.blur(),!0):!1)}Xt.prototype.observeUsing=function(e){this._observers===null&&(this._observers=new Set),this._observers.add(e),y(this._fragmentFiber.child,!1,p1,e,void 0,void 0)};function p1(e,t){return e.tag===6||(e=N(e),t.observe(e)),!1}Xt.prototype.unobserveUsing=function(e){var t=this._observers;if(t!==null&&t.has(e)){t.delete(e),y(this._fragmentFiber.child,!1,f1,e,void 0,void 0);for(var n=t=0;n<cn.length;n++){var s=cn[n];s.fragmentInstance===this&&s.observer===e?e.unobserve(s.instance):cn[t++]=s}cn.length=t}};function f1(e,t){return e.tag===6||(e=N(e),t.unobserve(e)),!1}var cn=[],Pc=!1;function m1(e,t,n){cn.push({fragmentInstance:e,observer:t,instance:n}),Pc||(Pc=!0,k1(function(){Pc=!1;var s=cn;cn=[];for(var i=0;i<s.length;i++){var l=s[i];l.observer.unobserve(l.instance)}}))}Xt.prototype.getClientRects=function(){var e=[];return y(this._fragmentFiber.child,!1,g1,e,void 0,void 0),e};function g1(e,t){if(e.tag===6){e=e.stateNode;var n=e.ownerDocument.createRange();n.selectNodeContents(e),t.push.apply(t,n.getClientRects())}else e=N(e),t.push.apply(t,e.getClientRects());return!1}Xt.prototype.getRootNode=function(e){var t=w(this._fragmentFiber);return t===null?this:N(t).getRootNode(e)},Xt.prototype.compareDocumentPosition=function(e){var t=w(this._fragmentFiber);if(t===null)return Node.DOCUMENT_POSITION_DISCONNECTED;var n=[];y(this._fragmentFiber.child,!1,Wc,n,void 0,void 0);var s=N(t);if(n.length===0){if(n=s,A(this._fragmentFiber)){e:{for(t=this._fragmentFiber.return;t!==null;){if(t.tag===4){t=t.stateNode.containerInfo;break e}if(t.tag===3||t.tag===5||t.tag===27)break;t=t.return}t=null}t!=null&&(n=t)}t=this._fragmentFiber;var i=s=n.compareDocumentPosition(e);return n===e?i=Node.DOCUMENT_POSITION_CONTAINS:s&Node.DOCUMENT_POSITION_CONTAINED_BY&&(n=E(t)[1],n===null?i=Node.DOCUMENT_POSITION_PRECEDING:(e=N(n).compareDocumentPosition(e),i=e===0||e&Node.DOCUMENT_POSITION_FOLLOWING?Node.DOCUMENT_POSITION_FOLLOWING:Node.DOCUMENT_POSITION_PRECEDING)),i|=Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC}t=N(n[0]),i=N(n[n.length-1]);var l=A(this._fragmentFiber)?t.parentElement:s;if(l==null)return Node.DOCUMENT_POSITION_DISCONNECTED;s=l.compareDocumentPosition(t)&Node.DOCUMENT_POSITION_CONTAINED_BY,l=l.compareDocumentPosition(i)&Node.DOCUMENT_POSITION_CONTAINED_BY;var d=t.compareDocumentPosition(e),p=i.compareDocumentPosition(e),k=d&Node.DOCUMENT_POSITION_CONTAINED_BY||p&Node.DOCUMENT_POSITION_CONTAINED_BY;return p=s&&l&&d&Node.DOCUMENT_POSITION_FOLLOWING&&p&Node.DOCUMENT_POSITION_PRECEDING,t=s&&t===e||l&&i===e||k||p?Node.DOCUMENT_POSITION_CONTAINED_BY:!s&&t===e||!l&&i===e?Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC:d,t&Node.DOCUMENT_POSITION_DISCONNECTED||t&Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC||v1(t,this._fragmentFiber,n[0],n[n.length-1],e)?t:Node.DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC};function v1(e,t,n,s,i){var l=ka(i);if(e&Node.DOCUMENT_POSITION_CONTAINED_BY){if(n=!!l)e:{for(;l!==null;){if(l.tag===7&&(l===t||l.alternate===t)){n=!0;break e}l=l.return}n=!1}return n}if(e&Node.DOCUMENT_POSITION_CONTAINS){if(l===null)return l=i.ownerDocument,i===l||i===l.documentElement||i===l.body;e:{for(l=t,t=w(t);l!==null;){if(!(l.tag!==5&&l.tag!==3&&l.tag!==27||l!==t&&l.alternate!==t)){l=!0;break e}l=l.return}l=!1}return l}return e&Node.DOCUMENT_POSITION_PRECEDING?((t=!!l)&&!(t=l===n)&&(t=S(n,l,Y),t===null?t=!1:(y(t,!0,G,l,n),l=D,D=null,t=l!==null)),t):e&Node.DOCUMENT_POSITION_FOLLOWING?((t=!!l)&&!(t=l===s)&&(t=S(s,l,Y),t===null?t=!1:(y(t,!0,Q,l,s),l=D,H=D=null,t=l!==null)),t):!1}function If(e,t){var n=e.ownerDocument.createRange();n.selectNodeContents(e),e=n.getBoundingClientRect(),window.scrollTo(window.scrollX+e.left,t?window.scrollY+e.top:window.scrollY+e.bottom-window.innerHeight)}Xt.prototype.scrollIntoView=function(e){if(typeof e=="object")throw Error(r(566));var t=[];y(this._fragmentFiber.child,!1,Wc,t,void 0,void 0);var n=e!==!1;if(t.length===0){var s=E(this._fragmentFiber);if(s=n?s[1]||s[0]||w(this._fragmentFiber):s[0]||s[1],s===null)return;if(s.tag===6){e=N(s),If(e,n);return}if(s=N(s),s.nodeType!==9){if(s.nodeType===11){n="host"in s?s.host:null,n!==null&&n.scrollIntoView(e);return}s.scrollIntoView(e)}}for(s=n?t.length-1:0;s!==(n?-1:t.length);){var i=t[s];i.tag===6?(i=N(i),If(i,n)):N(i).scrollIntoView(e),s+=n?-1:1}};function y1(e,t){return e=N(e),Gf(e,t),!1}function Gf(e,t){e.reactFragments==null&&(e.reactFragments=new Set),e.reactFragments.add(t)}function Bf(e,t){var n=t._eventListeners;if(n!==null)for(var s=0;s<n.length;s++){var i=n[s];e.addEventListener(i.type,i.attachedListener,Rs(i.optionsOrUseCapture))}e.nodeType!==3&&(n=t._observers,n!==null&&n.forEach(function(l){for(var d=0,p=0;p<cn.length;p++){var k=cn[p];(k.fragmentInstance!==t||k.observer!==l||k.instance!==e)&&(cn[d++]=k)}cn.length=d,l.observe(e)}),Gf(e,t))}function x1(e,t){var n=t._eventListeners;if(n!==null)for(var s=0;s<n.length;s++){var i=n[s];e.removeEventListener(i.type,i.attachedListener,Rs(i.optionsOrUseCapture))}e.nodeType!==3&&(n=t._observers,n!==null&&n.forEach(function(l){typeof l.rootMargin=="string"?m1(t,l,e):l.unobserve(e)}),e.reactFragments!=null&&e.reactFragments.delete(t))}function $c(e){var t=e.firstChild;for(t&&t.nodeType===10&&(t=t.nextSibling);t;){var n=t;switch(t=t.nextSibling,n.nodeName){case"HTML":case"HEAD":case"BODY":$c(n),lo(n);continue;case"SCRIPT":case"STYLE":continue;case"LINK":if(n.rel.toLowerCase()==="stylesheet")continue}e.removeChild(n)}}function w1(e,t,n,s){for(;e.nodeType===1;){var i=n;if(e.nodeName.toLowerCase()!==t.toLowerCase()){if(!s&&(e.nodeName!=="INPUT"||e.type!=="hidden"))break}else if(s){if(!e[ti])switch(t){case"meta":if(!e.hasAttribute("itemprop"))break;return e;case"link":if(l=e.getAttribute("rel"),l==="stylesheet"&&e.hasAttribute("data-precedence"))break;if(l!==i.rel||e.getAttribute("href")!==(i.href==null||i.href===""?null:i.href)||e.getAttribute("crossorigin")!==(i.crossOrigin==null?null:i.crossOrigin)||e.getAttribute("title")!==(i.title==null?null:i.title))break;return e;case"style":if(e.hasAttribute("data-precedence"))break;return e;case"script":if(l=e.getAttribute("src"),(l!==(i.src==null?null:i.src)||e.getAttribute("type")!==(i.type==null?null:i.type)||e.getAttribute("crossorigin")!==(i.crossOrigin==null?null:i.crossOrigin))&&l&&e.hasAttribute("async")&&!e.hasAttribute("itemprop"))break;return e;default:return e}}else if(t==="input"&&e.type==="hidden"){var l=i.name==null?null:""+i.name;if(i.type==="hidden"&&e.getAttribute("name")===l)return e}else return e;if(e=en(e.nextSibling),e===null)break}return null}function b1(e,t,n){if(t==="")return null;for(;e.nodeType!==3;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!n||(e=en(e.nextSibling),e===null))return null;return e}function Yf(e,t){for(;e.nodeType!==8;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!t||(e=en(e.nextSibling),e===null))return null;return e}function ed(e){return e.data==="$?"||e.data==="$~"}function td(e){return e.data==="$!"||e.data==="$?"&&e.ownerDocument.readyState!=="loading"}function j1(e,t){var n=e.ownerDocument;if(e.data==="$~")e._reactRetry=t;else if(e.data!=="$?"||n.readyState!=="loading")t();else{var s=function(){t(),n.removeEventListener("DOMContentLoaded",s)};n.addEventListener("DOMContentLoaded",s),e._reactRetry=s}}function en(e){for(;e!=null;e=e.nextSibling){var t=e.nodeType;if(t===1||t===3)break;if(t===8){if(t=e.data,t==="$"||t==="$!"||t==="$?"||t==="$~"||t==="&"||t==="F!"||t==="F")break;if(t==="/$"||t==="/&")return null}}return e}var nd=null;function Xf(e){e=e.nextSibling;for(var t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="/$"||n==="/&"){if(t===0)return en(e.nextSibling);t--}else n!=="$"&&n!=="$!"&&n!=="$?"&&n!=="$~"&&n!=="&"||t++}e=e.nextSibling}return null}function Vf(e){e=e.previousSibling;for(var t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="$"||n==="$!"||n==="$?"||n==="$~"||n==="&"){if(t===0)return e;t--}else n!=="/$"&&n!=="/&"||t++}e=e.previousSibling}return null}function A1(e,t){function n(){s=!0}if(e.ownerDocument.activeElement===e)return!0;var s=!1;try{e.ownerDocument.addEventListener("focus",n,!0),(e.focus||HTMLElement.prototype.focus).call(e,t)}finally{e.ownerDocument.removeEventListener("focus",n,!0)}return s}function k1(e){Of(function(){Of(function(t){return e(t)})})}function Qf(e,t,n){switch(t=Li(n),e){case"html":if(e=t.documentElement,!e)throw Error(r(452));return e;case"head":if(e=t.head,!e)throw Error(r(453));return e;case"body":if(e=t.body,!e)throw Error(r(454));return e;default:throw Error(r(451))}}function Zf(e,t,n){for(var s in n){var i=n[s];n.hasOwnProperty(s)&&i!=null&&Ue(e,t,s,null,$0,i)}n.dangerouslySetInnerHTML!=null&&(e.textContent=""),e.onclick===pn&&(e.onclick=null),lo(e)}function ad(e){for(var t=e.attributes;t.length;)e.removeAttributeNode(t[0]);lo(e)}var tn=new Map,Kf=new Set;function Ui(e){if(typeof e.getRootNode=="function"){var t=e.getRootNode();if(t.nodeType===9||t.nodeType===11)return t}return e.nodeType===9?e:e.ownerDocument}var Hn=he.d;he.d={f:C1,r:T1,D:S1,C:N1,L:_1,m:E1,X:z1,S:O1,M:R1};function C1(){var e=Hn.f(),t=ul();return e||t}function T1(e){var t=Pa(e);t!==null&&t.tag===5&&t.type==="form"?Fh(t):Hn.r(e)}var Ms=typeof document>"u"?null:document;function Ff(e,t,n){var s=Ms;if(s&&typeof t=="string"&&t){var i=Zt(t);i='link[rel="'+e+'"][href="'+i+'"]',typeof n=="string"&&(i+='[crossorigin="'+n+'"]'),Kf.has(i)||(Kf.add(i),e={rel:e,crossOrigin:n,href:t},s.querySelector(i)===null&&(t=s.createElement("link"),vt(t,"link",e),ot(t),s.head.appendChild(t)))}}function S1(e){Hn.D(e),Ff("dns-prefetch",e,null)}function N1(e,t){Hn.C(e,t),Ff("preconnect",e,t)}function _1(e,t,n){Hn.L(e,t,n);var s=Ms;if(s&&e&&t){var i='link[rel="preload"][as="'+Zt(t)+'"]';t==="image"&&n&&n.imageSrcSet?(i+='[imagesrcset="'+Zt(n.imageSrcSet)+'"]',typeof n.imageSizes=="string"&&(i+='[imagesizes="'+Zt(n.imageSizes)+'"]')):i+='[href="'+Zt(e)+'"]';var l=i;switch(t){case"style":l=qs(e);break;case"script":l=Ds(e)}if(!(tn.has(l)||(e=U({rel:"preload",href:t==="image"&&n&&n.imageSrcSet?void 0:e,as:t},n),tn.set(l,e),s.querySelector(i)!==null||t==="style"&&s.querySelector(Hi(l))||t==="script"&&s.querySelector(Ii(l))))){var d=s.createElement("link");vt(d,"link",e),t==="style"&&(d[oo]=!0,d.onload=d.onerror=function(){lu(d)}),ot(d),s.head.appendChild(d)}}}function E1(e,t){Hn.m(e,t);var n=Ms;if(n&&e){var s=t&&typeof t.as=="string"?t.as:"script",i='link[rel="modulepreload"][as="'+Zt(s)+'"][href="'+Zt(e)+'"]',l=i;switch(s){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":l=Ds(e)}if(!tn.has(l)&&(e=U({rel:"modulepreload",href:e},t),tn.set(l,e),n.querySelector(i)===null)){switch(s){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":if(n.querySelector(Ii(l)))return}s=n.createElement("link"),vt(s,"link",e),ot(s),n.head.appendChild(s)}}}function O1(e,t,n){Hn.S(e,t,n);var s=Ms;if(s&&e){var i=$a(s).hoistableStyles,l=qs(e);t=t||"default";var d=i.get(l);if(!d){var p={loading:0,preload:null};if(d=s.querySelector(Hi(l)))p.loading=5;else{e=U({rel:"stylesheet",href:e,"data-precedence":t},n),(n=tn.get(l))&&sd(e,n);var k=d=s.createElement("link");ot(k),vt(k,"link",e),k._p=new Promise(function(R,L){k.onload=R,k.onerror=L}),k.addEventListener("load",function(){p.loading|=1}),k.addEventListener("error",function(){p.loading|=2}),p.loading|=4,yl(d,t,s)}d={type:"stylesheet",instance:d,count:1,state:p},i.set(l,d)}}}function z1(e,t){Hn.X(e,t);var n=Ms;if(n&&e){var s=$a(n).hoistableScripts,i=Ds(e),l=s.get(i);l||(l=n.querySelector(Ii(i)),l||(e=U({src:e,async:!0},t),(t=tn.get(i))&&id(e,t),l=n.createElement("script"),ot(l),vt(l,"link",e),n.head.appendChild(l)),l={type:"script",instance:l,count:1,state:null},s.set(i,l))}}function R1(e,t){Hn.M(e,t);var n=Ms;if(n&&e){var s=$a(n).hoistableScripts,i=Ds(e),l=s.get(i);l||(l=n.querySelector(Ii(i)),l||(e=U({src:e,async:!0,type:"module"},t),(t=tn.get(i))&&id(e,t),l=n.createElement("script"),ot(l),vt(l,"link",e),n.head.appendChild(l)),l={type:"script",instance:l,count:1,state:null},s.set(i,l))}}function Jf(e,t,n,s){var i=(i=Fe.current)?Ui(i):null;if(!i)throw Error(r(446));switch(e){case"meta":case"title":return null;case"style":return typeof n.precedence=="string"&&typeof n.href=="string"?(n=qs(n.href),t=$a(i).hoistableStyles,s=t.get(n),s||(s={type:"style",instance:null,count:0,state:null},t.set(n,s)),s):{type:"void",instance:null,count:0,state:null};case"link":if(n.rel==="stylesheet"&&typeof n.href=="string"&&typeof n.precedence=="string"){e=qs(n.href);var l=$a(i).hoistableStyles,d=l.get(e);if(d||(i=i.ownerDocument||i,d={type:"stylesheet",instance:null,count:0,state:{loading:0,preload:null}},l.set(e,d),(l=i.querySelector(Hi(e)))?l._p||(d.instance=l,d.state.loading=5):(l=tn.get(e),l||(l={rel:"preload",as:"style",href:n.href,crossOrigin:n.crossOrigin,integrity:n.integrity,media:n.media,hrefLang:n.hrefLang,referrerPolicy:n.referrerPolicy},tn.set(e,l)),M1(i,e,l,d.state))),t&&s===null)throw Error(r(528,""));return d}if(t&&s!==null)throw Error(r(529,""));return null;case"script":return t=n.async,n=n.src,typeof n=="string"&&t&&typeof t!="function"&&typeof t!="symbol"?(n=Ds(n),t=$a(i).hoistableScripts,s=t.get(n),s||(s={type:"script",instance:null,count:0,state:null},t.set(n,s)),s):{type:"void",instance:null,count:0,state:null};default:throw Error(r(444,e))}}function qs(e){return'href="'+Zt(e)+'"'}function Hi(e){return'link[rel="stylesheet"]['+e+"]"}function Wf(e){return U({},e,{"data-precedence":e.precedence,precedence:null})}function M1(e,t,n,s){if(t=e.querySelector('link[rel="preload"][as="style"]['+t+"]")){if(t[oo]!==!0){s.loading=1;return}}else t=e.createElement("link"),t[oo]=!0,t.onload=t.onerror=lu.bind(null,t),vt(t,"link",n),ot(t),e.head.appendChild(t);s.preload=t,t.addEventListener("load",function(){return s.loading|=1}),t.addEventListener("error",function(){return s.loading|=2})}function Ds(e){return'[src="'+Zt(e)+'"]'}function Ii(e){return"script[async]"+e}function Pf(e,t,n){if(t.count++,t.instance===null)switch(t.type){case"style":var s=e.querySelector('style[data-href~="'+Zt(n.href)+'"]');if(s)return t.instance=s,ot(s),s;var i=U({},n,{"data-href":n.href,"data-precedence":n.precedence,href:null,precedence:null});return s=(e.ownerDocument||e).createElement("style"),ot(s),vt(s,"style",i),yl(s,n.precedence,e),t.instance=s;case"stylesheet":i=qs(n.href);var l=e.querySelector(Hi(i));if(l)return t.state.loading|=4,t.instance=l,ot(l),l;s=Wf(n),(i=tn.get(i))&&sd(s,i),l=(e.ownerDocument||e).createElement("link"),ot(l);var d=l;return d._p=new Promise(function(p,k){d.onload=p,d.onerror=k}),vt(l,"link",s),t.state.loading|=4,yl(l,n.precedence,e),t.instance=l;case"script":return l=Ds(n.src),(i=e.querySelector(Ii(l)))?(t.instance=i,ot(i),i):(s=n,(i=tn.get(l))&&(s=U({},n),id(s,i)),e=e.ownerDocument||e,i=e.createElement("script"),ot(i),vt(i,"link",s),e.head.appendChild(i),t.instance=i);case"void":return null;default:throw Error(r(443,t.type))}else t.type==="stylesheet"&&(t.state.loading&4)===0&&(s=t.instance,t.state.loading|=4,yl(s,n.precedence,e));return t.instance}function yl(e,t,n){for(var s=n.querySelectorAll('link[rel="stylesheet"][data-precedence],style[data-precedence]'),i=s.length?s[s.length-1]:null,l=i,d=0;d<s.length;d++){var p=s[d];if(p.dataset.precedence===t)l=p;else if(l!==i)break}l?l.parentNode.insertBefore(e,l.nextSibling):(t=n.nodeType===9?n.head:n,t.insertBefore(e,t.firstChild))}function sd(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.title==null&&(e.title=t.title)}function id(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.integrity==null&&(e.integrity=t.integrity)}var xl=null;function $f(e,t,n){if(xl===null){var s=new Map,i=xl=new Map;i.set(n,s)}else i=xl,s=i.get(n),s||(s=new Map,i.set(n,s));if(s.has(e))return s;for(s.set(e,null),n=n.getElementsByTagName(e),i=0;i<n.length;i++){var l=n[i];if(!(l[ti]||l[ht]||e==="link"&&l.getAttribute("rel")==="stylesheet")&&l.namespaceURI!=="http://www.w3.org/2000/svg"){var d=l.getAttribute(t)||"";d=e+d;var p=s.get(d);p?p.push(l):s.set(d,[l])}}return s}function od(e,t,n){e=e.ownerDocument||e,e.head.insertBefore(n,t==="title"?e.querySelector("head > title"):null)}function q1(e,t,n){if(n===1||t.itemProp!=null)return!1;switch(e){case"meta":case"title":return!0;case"style":if(typeof t.precedence!="string"||typeof t.href!="string"||t.href==="")break;return!0;case"link":if(typeof t.rel!="string"||typeof t.href!="string"||t.href===""||t.onLoad||t.onError)break;return t.rel==="stylesheet"?(e=t.disabled,typeof t.precedence=="string"&&e==null):!0;case"script":if(t.async&&typeof t.async!="function"&&typeof t.async!="symbol"&&!t.onLoad&&!t.onError&&t.src&&typeof t.src=="string")return!0}return!1}function em(e,t){return e==="img"&&t.src!=null&&t.src!==""&&t.onLoad==null&&t.loading!=="lazy"}function tm(e){return!(e.type==="stylesheet"&&(e.state.loading&3)===0)}function nm(e){return(e.width||100)*(e.height||100)*(typeof devicePixelRatio=="number"?devicePixelRatio:1)*.25}function am(e,t){typeof t.decode=="function"&&(e.imgCount++,t.complete||(e.imgBytes+=nm(t),e.suspenseyImages.push(t)),e=U1.bind(e),t.decode().then(e,e))}function D1(e,t,n,s){if(n.type==="stylesheet"&&(typeof s.media!="string"||matchMedia(s.media).matches!==!1)&&(n.state.loading&4)===0){if(n.instance===null){var i=qs(s.href),l=t.querySelector(Hi(i));if(l){t=l._p,t!==null&&typeof t=="object"&&typeof t.then=="function"&&(e.count++,e=Gi.bind(e),t.then(e,e)),n.state.loading|=4,n.instance=l,ot(l);return}l=t.ownerDocument||t,s=Wf(s),(i=tn.get(i))&&sd(s,i),l=l.createElement("link"),ot(l);var d=l;d._p=new Promise(function(p,k){d.onload=p,d.onerror=k}),vt(l,"link",s),n.instance=l}e.stylesheets===null&&(e.stylesheets=new Map),e.stylesheets.set(n,t),(t=n.state.preload)&&(n.state.loading&3)===0&&(e.count++,n=Gi.bind(e),t.addEventListener("load",n),t.addEventListener("error",n))}}var wl=0;function L1(e,t){return e.stylesheets&&e.count===0&&jl(e,e.stylesheets),0<e.count||0<e.imgCount?function(n){var s=setTimeout(function(){if(e.stylesheets&&jl(e,e.stylesheets),e.unsuspend){var l=e.unsuspend;e.unsuspend=null,l()}},6e4+t);0<e.imgBytes&&wl===0&&(wl=62500*t1());var i=setTimeout(function(){if(e.waitingForImages=!1,e.count===0&&(e.stylesheets&&jl(e,e.stylesheets),e.unsuspend)){var l=e.unsuspend;e.unsuspend=null,l()}},(e.imgBytes>wl?50:800)+t);return e.unsuspend=n,function(){e.unsuspend=null,clearTimeout(s),clearTimeout(i)}}:null}function sm(e){if(e.count===0&&(e.imgCount===0||!e.waitingForImages)){if(e.stylesheets)jl(e,e.stylesheets);else if(e.unsuspend){var t=e.unsuspend;e.unsuspend=null,t()}}}function Gi(){this.count--,sm(this)}function U1(){this.imgCount--,sm(this)}var bl=null;function jl(e,t){e.stylesheets=null,e.unsuspend!==null&&(e.count++,bl=new Map,t.forEach(H1,e),bl=null,Gi.call(e))}function H1(e,t){if(!(t.state.loading&4)){var n=bl.get(e);if(n)var s=n.get(null);else{n=new Map,bl.set(e,n);for(var i=e.querySelectorAll("link[data-precedence],style[data-precedence]"),l=0;l<i.length;l++){var d=i[l];(d.nodeName==="LINK"||d.getAttribute("media")!=="not all")&&(n.set(d.dataset.precedence,d),s=d)}s&&n.set(null,s)}i=t.instance,d=i.getAttribute("data-precedence"),l=n.get(d)||s,l===s&&n.set(null,i),n.set(d,i),this.count++,s=Gi.bind(this),i.addEventListener("load",s),i.addEventListener("error",s),l?l.parentNode.insertBefore(i,l.nextSibling):(e=e.nodeType===9?e.head:e,e.insertBefore(i,e.firstChild)),t.state.loading|=4}}var Ls={$$typeof:W,Provider:null,Consumer:null,_currentValue:ut,_currentValue2:ut,_threadCount:0};function I1(e,t,n,s,i,l,d,p,k){this.tag=1,this.containerInfo=e,this.pingCache=this.current=this.pendingChildren=null,this.timeoutHandle=-1,this.callbackNode=this.next=this.pendingContext=this.context=this.cancelPendingCommit=null,this.callbackPriority=0,this.expirationTimes=Il(-1),this.entangledLanes=this.shellSuspendCounter=this.errorRecoveryDisabledLanes=this.expiredLanes=this.warmLanes=this.pingedLanes=this.suspendedLanes=this.pendingLanes=0,this.entanglements=Il(0),this.hiddenUpdates=Il(null),this.identifierPrefix=s,this.onUncaughtError=i,this.onCaughtError=l,this.onRecoverableError=d,this.pooledCache=null,this.pooledCacheLanes=0,this.formState=k,this.transitionTypes=null,this.incompleteTransitions=new Map}function im(e,t,n,s,i,l,d,p,k,R,L,V){return e=new I1(e,t,n,d,k,R,L,V,p),t=1,l===!0&&(t|=24),l=St(3,null,null,t),e.current=l,l.stateNode=e,t=br(),t.refCount++,e.pooledCache=t,t.refCount++,l.memoizedState={element:s,isDehydrated:n,cache:t},Cr(l),e}function om(e){return e?(e=cs,e):cs}function lm(e,t,n,s,i,l){i=om(i),s.context===null?s.context=i:s.pendingContext=i,s=Pn(t),s.payload={element:n},l=l===void 0?null:l,l!==null&&(s.callback=l),n=$n(e,s,t),n!==null&&(Ot(n,e,t),yi(n,e,t))}function rm(e,t){if(e=e.memoizedState,e!==null&&e.dehydrated!==null){var n=e.retryLane;e.retryLane=n!==0&&n<t?n:t}}function ld(e,t){rm(e,t),(e=e.alternate)&&rm(e,t)}function cm(e){if(e.tag===13||e.tag===31){var t=Na(e,67108864);t!==null&&Ot(t,e,67108864),ld(e,67108864)}}function dm(e){if(e.tag===13||e.tag===31){var t=Yt();t=Gl(t);var n=Na(e,t);n!==null&&Ot(n,e,t),ld(e,t)}}var Us=!0;function G1(e,t,n,s){var i=ae.T;ae.T=null;var l=he.p;try{he.p=2,rd(e,t,n,s)}finally{he.p=l,ae.T=i}}function B1(e,t,n,s){var i=ae.T;ae.T=null;var l=he.p;try{he.p=8,rd(e,t,n,s)}finally{he.p=l,ae.T=i}}function rd(e,t,n,s){if(Us){var i=cd(s);if(i===null)Yc(e,t,s,Al,n),hm(e,s);else if(X1(i,e,t,n,s))s.stopPropagation();else if(hm(e,s),t&4&&-1<Y1.indexOf(e)){for(;i!==null;){var l=Pa(i);if(l!==null)switch(l.tag){case 3:if(l=l.stateNode,l.current.memoizedState.isDehydrated){var d=Aa(l.pendingLanes);if(d!==0){var p=l;for(p.pendingLanes|=2,p.entangledLanes|=2;d;){var k=1<<31-Dt(d);p.entanglements[1]|=k,d&=~k}An(l),(Me&6)===0&&(rl=Mt()+500,Mi(0))}}break;case 31:case 13:p=Na(l,2),p!==null&&Ot(p,l,2),ul(),ld(l,2)}if(l=cd(s),l===null&&Yc(e,t,s,Al,n),l===i)break;i=l}i!==null&&s.stopPropagation()}else Yc(e,t,s,null,n)}}function cd(e){return e=Kl(e),dd(e)}var Al=null;function dd(e){if(Al=null,e=ka(e),e!==null){var t=m(e);if(t===null)e=null;else{var n=t.tag;if(n===13){if(e=g(t),e!==null)return e;e=null}else if(n===31){if(e=v(t),e!==null)return e;e=null}else if(n===3){if(t.stateNode.current.memoizedState.isDehydrated)return t.tag===3?t.stateNode.containerInfo:null;e=null}else t!==e&&(e=null)}}return Al=e,null}function um(e){switch(e){case"beforetoggle":case"cancel":case"click":case"close":case"contextmenu":case"copy":case"cut":case"auxclick":case"dblclick":case"dragend":case"dragstart":case"drop":case"focusin":case"focusout":case"input":case"invalid":case"keydown":case"keypress":case"keyup":case"mousedown":case"mouseup":case"paste":case"pause":case"play":case"pointercancel":case"pointerdown":case"pointerup":case"ratechange":case"reset":case"seeked":case"submit":case"toggle":case"touchcancel":case"touchend":case"touchstart":case"volumechange":case"change":case"selectionchange":case"textInput":case"compositionstart":case"compositionend":case"compositionupdate":case"beforeblur":case"afterblur":case"beforeinput":case"blur":case"fullscreenchange":case"fullscreenerror":case"focus":case"hashchange":case"popstate":case"select":case"selectstart":return 2;case"drag":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"mousemove":case"mouseout":case"mouseover":case"pointermove":case"pointerout":case"pointerover":case"resize":case"scroll":case"touchmove":case"wheel":case"mouseenter":case"mouseleave":case"pointerenter":case"pointerleave":return 8;case"message":switch(ty()){case Fd:return 2;case Jd:return 8;case to:case ny:return 32;case Wd:return 268435456;default:return 32}default:return 32}}var ud=!1,ua=null,ha=null,pa=null,Bi=new Map,Yi=new Map,fa=[],Y1="mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(" ");function hm(e,t){switch(e){case"focusin":case"focusout":ua=null;break;case"dragenter":case"dragleave":ha=null;break;case"mouseover":case"mouseout":pa=null;break;case"pointerover":case"pointerout":Bi.delete(t.pointerId);break;case"gotpointercapture":case"lostpointercapture":Yi.delete(t.pointerId)}}function Xi(e,t,n,s,i,l){return e===null||e.nativeEvent!==l?(e={blockedOn:t,domEventName:n,eventSystemFlags:s,nativeEvent:l,targetContainers:[i]},t!==null&&(t=Pa(t),t!==null&&cm(t)),e):(e.eventSystemFlags|=s,t=e.targetContainers,i!==null&&t.indexOf(i)===-1&&t.push(i),e)}function X1(e,t,n,s,i){switch(t){case"focusin":return ua=Xi(ua,e,t,n,s,i),!0;case"dragenter":return ha=Xi(ha,e,t,n,s,i),!0;case"mouseover":return pa=Xi(pa,e,t,n,s,i),!0;case"pointerover":var l=i.pointerId;return Bi.set(l,Xi(Bi.get(l)||null,e,t,n,s,i)),!0;case"gotpointercapture":return l=i.pointerId,Yi.set(l,Xi(Yi.get(l)||null,e,t,n,s,i)),!0}return!1}function pm(e){var t=ka(e.target);if(t!==null){var n=m(t);if(n!==null){if(t=n.tag,t===13){if(t=g(n),t!==null){e.blockedOn=t,su(e.priority,function(){dm(n)});return}}else if(t===31){if(t=v(n),t!==null){e.blockedOn=t,su(e.priority,function(){dm(n)});return}}else if(t===3&&n.stateNode.current.memoizedState.isDehydrated){e.blockedOn=n.tag===3?n.stateNode.containerInfo:null;return}}}e.blockedOn=null}function kl(e){if(e.blockedOn!==null)return!1;for(var t=e.targetContainers;0<t.length;){var n=cd(e.nativeEvent);if(n===null){n=e.nativeEvent;var s=new n.constructor(n.type,n);Zl=s,n.target.dispatchEvent(s),Zl=null}else return t=Pa(n),t!==null&&cm(t),e.blockedOn=n,!1;t.shift()}return!0}function fm(e,t,n){kl(e)&&n.delete(t)}function V1(){ud=!1,ua!==null&&kl(ua)&&(ua=null),ha!==null&&kl(ha)&&(ha=null),pa!==null&&kl(pa)&&(pa=null),Bi.forEach(fm),Yi.forEach(fm)}function Cl(e,t){e.blockedOn===t&&(e.blockedOn=null,ud||(ud=!0,o.unstable_scheduleCallback(o.unstable_NormalPriority,V1)))}var Tl=null;function mm(e){Tl!==e&&(Tl=e,o.unstable_scheduleCallback(o.unstable_NormalPriority,function(){Tl===e&&(Tl=null);for(var t=0;t<e.length;t+=3){var n=e[t],s=e[t+1],i=e[t+2];if(typeof s!="function"){if(dd(s||n)===null)continue;break}var l=Pa(n);l!==null&&(e.splice(t,3),t-=3,Qr(l,{pending:!0,data:i,method:n.method,action:s},s,i))}}))}function Hs(e){function t(k){return Cl(k,e)}ua!==null&&Cl(ua,e),ha!==null&&Cl(ha,e),pa!==null&&Cl(pa,e),Bi.forEach(t),Yi.forEach(t);for(var n=0;n<fa.length;n++){var s=fa[n];s.blockedOn===e&&(s.blockedOn=null)}for(;0<fa.length&&(n=fa[0],n.blockedOn===null);)pm(n),n.blockedOn===null&&fa.shift();if(n=(e.ownerDocument||e).$$reactFormReplay,n!=null)for(s=0;s<n.length;s+=3){var i=n[s],l=n[s+1],d=i[Tt]||null;if(typeof l=="function")d||mm(n);else if(d){var p=null;if(l&&l.hasAttribute("formAction")){if(i=l,d=l[Tt]||null)p=d.formAction;else if(dd(i)!==null)continue}else p=d.action;typeof p=="function"?n[s+1]=p:(n.splice(s,3),s-=3),mm(n)}}}function gm(){function e(l){l.canIntercept&&l.info==="react-transition"&&l.intercept({handler:function(){return new Promise(function(d){return i=d})},focusReset:"manual",scroll:"manual"})}function t(){i!==null&&(i(),i=null),s||setTimeout(n,20)}function n(){if(!s&&!navigation.transition){var l=navigation.currentEntry;l&&l.url!=null&&navigation.navigate(l.url,{state:l.getState(),info:"react-transition",history:"replace"})}}if(typeof navigation=="object"){var s=!1,i=null;return navigation.addEventListener("navigate",e),navigation.addEventListener("navigatesuccess",t),navigation.addEventListener("navigateerror",t),setTimeout(n,100),function(){s=!0,navigation.removeEventListener("navigate",e),navigation.removeEventListener("navigatesuccess",t),navigation.removeEventListener("navigateerror",t),i!==null&&(i(),i=null)}}}function hd(e){this._internalRoot=e}Sl.prototype.render=hd.prototype.render=function(e){var t=this._internalRoot;if(t===null)throw Error(r(409));var n=t.current,s=Yt();lm(n,s,e,t,null,null)},Sl.prototype.unmount=hd.prototype.unmount=function(){var e=this._internalRoot;if(e!==null){this._internalRoot=null;var t=e.containerInfo;lm(e.current,2,null,e,null,null),ul(),t[Wa]=null}};function Sl(e){this._internalRoot=e}Sl.prototype.unstable_scheduleHydration=function(e){if(e){var t=au();e={blockedOn:null,target:e,priority:t};for(var n=0;n<fa.length&&t!==0&&t<fa[n].priority;n++);fa.splice(n,0,e),n===0&&pm(e)}};var vm=c.version;if(vm!=="19.3.0")throw Error(r(527,vm,"19.3.0"));he.findDOMNode=function(e){var t=e._reactInternals;if(t===void 0)throw typeof e.render=="function"?Error(r(188)):(e=Object.keys(e).join(","),Error(r(268,e)));return e=C(t),e=e!==null?x(e):null,e=e===null?null:e.stateNode,e};var Q1={bundleType:0,version:"19.3.0",rendererPackageName:"react-dom",currentDispatcherRef:ae,reconcilerVersion:"19.3.0"};if(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__<"u"){var Nl=__REACT_DEVTOOLS_GLOBAL_HOOK__;if(!Nl.isDisabled&&Nl.supportsFiber)try{Ps=Nl.inject(Q1),qt=Nl}catch{}}return Qi.createRoot=function(e,t){if(!h(e))throw Error(r(299));var n=!1,s="",i=ip,l=op,d=lp;return t!=null&&(t.unstable_strictMode===!0&&(n=!0),t.identifierPrefix!==void 0&&(s=t.identifierPrefix),t.onUncaughtError!==void 0&&(i=t.onUncaughtError),t.onCaughtError!==void 0&&(l=t.onCaughtError),t.onRecoverableError!==void 0&&(d=t.onRecoverableError)),t=im(e,1,!1,null,null,n,s,null,i,l,d,gm),e[Wa]=t.current,Bc(e),new hd(t)},Qi.hydrateRoot=function(e,t,n){if(!h(e))throw Error(r(299));var s=!1,i="",l=ip,d=op,p=lp,k=null;return n!=null&&(n.unstable_strictMode===!0&&(s=!0),n.identifierPrefix!==void 0&&(i=n.identifierPrefix),n.onUncaughtError!==void 0&&(l=n.onUncaughtError),n.onCaughtError!==void 0&&(d=n.onCaughtError),n.onRecoverableError!==void 0&&(p=n.onRecoverableError),n.formState!==void 0&&(k=n.formState)),t=im(e,1,!0,t,n??null,s,i,k,l,d,p,gm),t.context=om(null),n=t.current,s=Yt(),s=Gl(s),i=Pn(s),i.callback=null,$n(n,i,s),n=s,t.current.lanes=n,ei(t,n),An(t),e[Wa]=t.current,Bc(e),new Sl(t)},Qi.version="19.3.0",Qi}var Sm;function ax(){if(Sm)return md.exports;Sm=1;function o(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__>"u"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(o)}catch(c){console.error(c)}}return o(),md.exports=nx(),md.exports}var sx=ax();const ix=o=>o?.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase();function ox(o,c,u=[]){if(c==null)throw new Error("[lucide]: iconNode is required when icon name is used");return{name:ix(o),size:24,node:c,...u.length>0?{aliases:u}:{}}}const lx=o=>{let c="",u=!1;for(const r of o){if(r==="-"||r==="_"||r<=" "){u=c.length>0;continue}c.length===0?c+=r.toLowerCase():c+=u?r.toUpperCase():r,u=!1}return c};const rx=o=>{const c=lx(o);return c.charAt(0).toUpperCase()+c.slice(1)};const Td=(...o)=>o.filter((c,u,r)=>!!c&&c.trim()!==""&&r.indexOf(c)===u).join(" ").trim();const Qa={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":2,"stroke-linecap":"round","stroke-linejoin":"round"};function xd(o){return o!=null}function cx(o,c={}){const u=c.attributeNames??{},r=w=>u[w]??w,h=o.size??o.width??Qa.width,m=o.size??o.height??Qa.height,g=o.aliases?.filter(w=>typeof w=="string"&&w.trim()!=="").map(w=>`lucide-${w}`)??[],v=[...o.name?[`lucide-${o.name}`]:[],...g],b=c.className?.split(" ").filter(Boolean)??[],C=c.includeDefaultClasses===!1?Td(...b):Td("lucide",...v,...b),x=c.absoluteStrokeWidth?Number(c.strokeWidth??Qa["stroke-width"])*Number(o.size??o.width??Qa.width)/Number(c.size??c.width??Qa.width):c.strokeWidth??Qa["stroke-width"];return["svg",{...Object.entries(Qa).reduce((w,[A,E])=>(w[r(A)]=E,w),{}),..."color"in c&&c.color&&{[r("stroke")]:c.color},..."size"in c&&xd(c.size)&&{[r("width")]:c.size,[r("height")]:c.size},..."width"in c&&xd(c.width)&&{[r("width")]:c.width},..."height"in c&&xd(c.height)&&{[r("height")]:c.height},[r("stroke-width")]:x,...C&&{[r("class")]:C},[r("viewBox")]:`0 0 ${h} ${m}`,...c.hasA11yProp===!1?{[r("aria-hidden")]:"true"}:{},..."attributes"in c&&c.attributes},o.node.map(w=>{const[A,E,_]=w,N=c.nonScalingStroke?{[r("vector-effect")]:"non-scaling-stroke",...E}:E;return _?[A,N,_]:[A,N]})]}function dx(o,c={}){return cx(o,{...c,attributeNames:{...c.attributeNames,class:"className","stroke-width":"strokeWidth","stroke-linecap":"strokeLinecap","stroke-linejoin":"strokeLinejoin","vector-effect":"vectorEffect"}})}const ux=o=>{for(const c in o)if(c.startsWith("aria-")||c==="role"||c==="title")return!0;return!1},hx=f.createContext({}),px=()=>f.useContext(hx),fx=f.forwardRef(({color:o,size:c,width:u,height:r,strokeWidth:h,absoluteStrokeWidth:m,nonScalingStroke:g,className:v="",children:b,iconNode:C=[],icon:x={node:C,aliases:[],size:24},...y},w)=>{const{size:A=24,strokeWidth:E=2,absoluteStrokeWidth:_=!1,nonScalingStroke:N=!1,color:D="currentColor",className:H=""}=px()??{},G=!!b||ux(y),[Q,Y,S=[]]=dx(x,{color:o??D,width:u??c??A,height:r??c??A,strokeWidth:h??E,absoluteStrokeWidth:m??_,nonScalingStroke:g??N,className:Td(H,v),hasA11yProp:G,attributes:y});return f.createElement(Q,{ref:w,...Y},[...S.map(([U,Z])=>f.createElement(U,Z)),...Array.isArray(b)?b:[b]])});function ce(o,c=[],u=[]){const r=typeof o=="string"?ox(o,c,u):o,h=f.forwardRef(({className:m,...g},v)=>f.createElement(fx,{ref:v,icon:r,className:m,...g}));return r.name&&(h.displayName=rx(r.name)),h}const Fm={name:"activity",size:24,node:[["path",{d:"M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2",key:"169zse"}]]};Fm.node;const Jm=ce(Fm);const Wm={name:"app-window",size:24,node:[["rect",{x:"2",y:"4",width:"20",height:"16",rx:"2",key:"izxlao"}],["path",{d:"M10 4v4",key:"pp8u80"}],["path",{d:"M2 8h20",key:"d11cs7"}],["path",{d:"M6 4v4",key:"1svtjw"}]]};Wm.node;const Pm=ce(Wm);const $m={name:"arrow-right",size:24,node:[["path",{d:"M5 12h14",key:"1ays0h"}],["path",{d:"m12 5 7 7-7 7",key:"xquz4c"}]]};$m.node;const Ld=ce($m);const eg={name:"arrow-up-right",size:24,node:[["path",{d:"M7 7h10v10",key:"1tivn9"}],["path",{d:"M7 17 17 7",key:"1vkiza"}]]};eg.node;const dn=ce(eg);const tg={name:"bell-off",size:24,node:[["path",{d:"M10.268 21a2 2 0 0 0 3.464 0",key:"vwvbt9"}],["path",{d:"M17 17H4a1 1 0 0 1-.74-1.673C4.59 13.956 6 12.499 6 8a6 6 0 0 1 .258-1.742",key:"178tsu"}],["path",{d:"m2 2 20 20",key:"1ooewy"}],["path",{d:"M8.668 3.01A6 6 0 0 1 18 8c0 2.687.77 4.653 1.707 6.05",key:"1hqiys"}]]};tg.node;const mx=ce(tg);const ng={name:"bell",size:24,node:[["path",{d:"M10.268 21a2 2 0 0 0 3.464 0",key:"vwvbt9"}],["path",{d:"M3.262 15.326A1 1 0 0 0 4 17h16a1 1 0 0 0 .74-1.673C19.41 13.956 18 12.499 18 8A6 6 0 0 0 6 8c0 4.499-1.411 5.956-2.738 7.326",key:"11g9vi"}]]};ng.node;const gx=ce(ng);const ag={name:"bot",size:24,node:[["path",{d:"M12 8V4H8",key:"hb8ula"}],["rect",{width:"16",height:"12",x:"4",y:"8",rx:"2",key:"enze0r"}],["path",{d:"M2 14h2",key:"vft8re"}],["path",{d:"M20 14h2",key:"4cs60a"}],["path",{d:"M15 13v2",key:"1xurst"}],["path",{d:"M9 13v2",key:"rq6x2g"}]]};ag.node;const Ud=ce(ag);const sg={name:"boxes",size:24,node:[["path",{d:"M2.97 12.92A2 2 0 0 0 2 14.63v3.24a2 2 0 0 0 .97 1.71l3 1.8a2 2 0 0 0 2.06 0L12 19v-5.5l-5-3-4.03 2.42Z",key:"lc1i9w"}],["path",{d:"m7 16.5-4.74-2.85",key:"1o9zyk"}],["path",{d:"m7 16.5 5-3",key:"va8pkn"}],["path",{d:"M7 16.5v5.17",key:"jnp8gn"}],["path",{d:"M12 13.5V19l3.97 2.38a2 2 0 0 0 2.06 0l3-1.8a2 2 0 0 0 .97-1.71v-3.24a2 2 0 0 0-.97-1.71L17 10.5l-5 3Z",key:"8zsnat"}],["path",{d:"m17 16.5-5-3",key:"8arw3v"}],["path",{d:"m17 16.5 4.74-2.85",key:"8rfmw"}],["path",{d:"M17 16.5v5.17",key:"k6z78m"}],["path",{d:"M7.97 4.42A2 2 0 0 0 7 6.13v4.37l5 3 5-3V6.13a2 2 0 0 0-.97-1.71l-3-1.8a2 2 0 0 0-2.06 0l-3 1.8Z",key:"1xygjf"}],["path",{d:"M12 8 7.26 5.15",key:"1vbdud"}],["path",{d:"m12 8 4.74-2.85",key:"3rx089"}],["path",{d:"M12 13.5V8",key:"1io7kd"}]]};sg.node;const vx=ce(sg);const ig={name:"check-check",size:24,node:[["path",{d:"M18 6 7 17l-5-5",key:"116fxf"}],["path",{d:"m22 10-7.5 7.5L13 16",key:"ke71qq"}]]};ig.node;const yx=ce(ig);const og={name:"check",size:24,node:[["path",{d:"M20 6 9 17l-5-5",key:"1gmf2c"}]]};og.node;const Zs=ce(og);const lg={name:"chevron-down",size:24,node:[["path",{d:"m6 9 6 6 6-6",key:"qrunsl"}]]};lg.node;const xx=ce(lg);const rg={name:"chevron-right",size:24,node:[["path",{d:"m9 18 6-6-6-6",key:"mthhwq"}]]};rg.node;const un=ce(rg);const cg={name:"circle-alert",size:24,node:[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12",key:"1pkeuh"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16",key:"4dfq90"}]],aliases:["alert-circle"]};cg.node;const wx=ce(cg);const dg={name:"circle-check",size:24,node:[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["path",{d:"m16 9-5.5 5.5L8 12",key:"xofnsj"}]],aliases:["check-circle-2"]};dg.node;const ug=ce(dg);const hg={name:"circle-dashed",size:24,node:[["path",{d:"M10.1 2.182a10 10 0 0 1 3.8 0",key:"5ilxe3"}],["path",{d:"M13.9 21.818a10 10 0 0 1-3.8 0",key:"11zvb9"}],["path",{d:"M17.609 3.721a10 10 0 0 1 2.69 2.7",key:"1iw5b2"}],["path",{d:"M2.182 13.9a10 10 0 0 1 0-3.8",key:"c0bmvh"}],["path",{d:"M20.279 17.609a10 10 0 0 1-2.7 2.69",key:"1ruxm7"}],["path",{d:"M21.818 10.1a10 10 0 0 1 0 3.8",key:"qkgqxc"}],["path",{d:"M3.721 6.391a10 10 0 0 1 2.7-2.69",key:"1mcia2"}],["path",{d:"M6.391 20.279a10 10 0 0 1-2.69-2.7",key:"1fvljs"}]]};hg.node;const bx=ce(hg);const pg={name:"circle-question-mark",size:24,node:[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["path",{d:"M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3",key:"1u773s"}],["path",{d:"M12 17h.01",key:"p32p05"}]],aliases:["help-circle","circle-help"]};pg.node;const jx=ce(pg);const fg={name:"cloud",size:24,node:[["path",{d:"M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z",key:"p7xjir"}]]};fg.node;const mg=ce(fg);const gg={name:"copy",size:24,node:[["rect",{width:"14",height:"14",x:"8",y:"8",rx:"2",ry:"2",key:"17jyea"}],["path",{d:"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2",key:"zix9uf"}]]};gg.node;const Fa=ce(gg);const vg={name:"cpu",size:24,node:[["path",{d:"M12 20v2",key:"1lh1kg"}],["path",{d:"M12 2v2",key:"tus03m"}],["path",{d:"M17 20v2",key:"1rnc9c"}],["path",{d:"M17 2v2",key:"11trls"}],["path",{d:"M2 12h2",key:"1t8f8n"}],["path",{d:"M2 17h2",key:"7oei6x"}],["path",{d:"M2 7h2",key:"asdhe0"}],["path",{d:"M20 12h2",key:"1q8mjw"}],["path",{d:"M20 17h2",key:"1fpfkl"}],["path",{d:"M20 7h2",key:"1o8tra"}],["path",{d:"M7 20v2",key:"4gnj0m"}],["path",{d:"M7 2v2",key:"1i4yhu"}],["rect",{x:"4",y:"4",width:"16",height:"16",rx:"2",key:"1vbyd7"}],["rect",{x:"8",y:"8",width:"8",height:"8",rx:"1",key:"z9xiuo"}]]};vg.node;const Ax=ce(vg);const yg={name:"external-link",size:24,node:[["path",{d:"M15 3h6v6",key:"1q9fwt"}],["path",{d:"M10 14 21 3",key:"gplh6r"}],["path",{d:"M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6",key:"a6xqqp"}]]};yg.node;const Sd=ce(yg);const xg={name:"file-check-corner",size:24,node:[["path",{d:"M10.5 22H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.706.706l3.588 3.588A2.4 2.4 0 0 1 20 8v6",key:"g5mvt7"}],["path",{d:"M14 2v5a1 1 0 0 0 1 1h5",key:"wfsgrz"}],["path",{d:"m14 20 2 2 4-4",key:"15kota"}]],aliases:["file-check-2"]};xg.node;const kx=ce(xg);const wg={name:"hard-drive",size:24,node:[["path",{d:"M10 16h.01",key:"1bzywj"}],["path",{d:"M2.212 11.577a2 2 0 0 0-.212.896V18a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-5.527a2 2 0 0 0-.212-.896L18.55 5.11A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z",key:"18tbho"}],["path",{d:"M21.946 12.013H2.054",key:"zqlbp7"}],["path",{d:"M6 16h.01",key:"1pmjb7"}]]};wg.node;const Cx=ce(wg);const bg={name:"inbox",size:24,node:[["polyline",{points:"22 12 16 12 14 15 10 15 8 12 2 12",key:"o97t9d"}],["path",{d:"M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z",key:"oot6mr"}]]};bg.node;const Tx=ce(bg);const jg={name:"info",size:24,node:[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["path",{d:"M12 16v-4",key:"1dtifu"}],["path",{d:"M12 8h.01",key:"e9boi3"}]]};jg.node;const Sx=ce(jg);const Ag={name:"key-round",size:24,node:[["path",{d:"M2.586 17.414A2 2 0 0 0 2 18.828V21a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h1a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h.172a2 2 0 0 0 1.414-.586l.814-.814a6.5 6.5 0 1 0-4-4z",key:"1s6t7t"}],["circle",{cx:"16.5",cy:"7.5",r:".5",fill:"currentColor",key:"w0ekpg"}]]};Ag.node;const zl=ce(Ag);const kg={name:"layers",size:24,node:[["path",{d:"M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z",key:"zw3jo"}],["path",{d:"M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12",key:"1wduqc"}],["path",{d:"M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17",key:"kqbvx6"}]],aliases:["layers-3"]};kg.node;const Nx=ce(kg);const Cg={name:"layout-dashboard",size:24,node:[["rect",{width:"7",height:"9",x:"3",y:"3",rx:"1",key:"10lvy0"}],["rect",{width:"7",height:"5",x:"14",y:"3",rx:"1",key:"16une8"}],["rect",{width:"7",height:"9",x:"14",y:"12",rx:"1",key:"1hutg5"}],["rect",{width:"7",height:"5",x:"3",y:"16",rx:"1",key:"ldoo1y"}]]};Cg.node;const _x=ce(Cg);const Tg={name:"lightbulb",size:24,node:[["path",{d:"M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5",key:"1gvzjb"}],["path",{d:"M9 18h6",key:"x1upvd"}],["path",{d:"M10 22h4",key:"ceow96"}]]};Tg.node;const Ex=ce(Tg);const Sg={name:"list-todo",size:24,node:[["path",{d:"M13 5h8",key:"a7qcls"}],["path",{d:"M13 12h8",key:"h98zly"}],["path",{d:"M13 19h8",key:"c3s6r1"}],["path",{d:"m3 17 2 2 4-4",key:"1jhpwq"}],["rect",{x:"3",y:"4",width:"6",height:"6",rx:"1",key:"cif1o7"}]]};Sg.node;const Ng=ce(Sg);const _g={name:"log-out",size:24,node:[["path",{d:"m16 17 5-5-5-5",key:"1bji2h"}],["path",{d:"M21 12H9",key:"dn1m92"}],["path",{d:"M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4",key:"1uf3rs"}]]};_g.node;const Ox=ce(_g);const Eg={name:"monitor",size:24,node:[["rect",{width:"20",height:"14",x:"2",y:"3",rx:"2",key:"48i651"}],["line",{x1:"8",x2:"16",y1:"21",y2:"21",key:"1svkeh"}],["line",{x1:"12",x2:"12",y1:"17",y2:"21",key:"vw1qmm"}]]};Eg.node;const zx=ce(Eg);const Og={name:"moon",size:24,node:[["path",{d:"M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401",key:"kfwtm"}]]};Og.node;const Rx=ce(Og);const zg={name:"panel-left-close",size:24,node:[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2",key:"afitv7"}],["path",{d:"M9 3v18",key:"fh3hqa"}],["path",{d:"m16 15-3-3 3-3",key:"14y99z"}]],aliases:["sidebar-close"]};zg.node;const Mx=ce(zg);const Rg={name:"panel-left-open",size:24,node:[["rect",{width:"18",height:"18",x:"3",y:"3",rx:"2",key:"afitv7"}],["path",{d:"M9 3v18",key:"fh3hqa"}],["path",{d:"m14 9 3 3-3 3",key:"8010ee"}]],aliases:["sidebar-open"]};Rg.node;const qx=ce(Rg);const Mg={name:"pencil",size:24,node:[["path",{d:"M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z",key:"1a8usu"}],["path",{d:"m15 5 4 4",key:"1mk7zo"}]]};Mg.node;const Nd=ce(Mg);const qg={name:"plug",size:24,node:[["path",{d:"M12 22v-5",key:"1ega77"}],["path",{d:"M15 8V2",key:"18g5xt"}],["path",{d:"M17 8a1 1 0 0 1 1 1v4a4 4 0 0 1-4 4h-4a4 4 0 0 1-4-4V9a1 1 0 0 1 1-1z",key:"1xoxul"}],["path",{d:"M9 8V2",key:"14iosj"}]]};qg.node;const Dx=ce(qg);const Dg={name:"plus",size:24,node:[["path",{d:"M5 12h14",key:"1ays0h"}],["path",{d:"M12 5v14",key:"s699le"}]]};Dg.node;const ql=ce(Dg);const Lg={name:"puzzle",size:24,node:[["path",{d:"M15.39 4.39a1 1 0 0 0 1.68-.474 2.5 2.5 0 1 1 3.014 3.015 1 1 0 0 0-.474 1.68l1.683 1.682a2.414 2.414 0 0 1 0 3.414L19.61 15.39a1 1 0 0 1-1.68-.474 2.5 2.5 0 1 0-3.014 3.015 1 1 0 0 1 .474 1.68l-1.683 1.682a2.414 2.414 0 0 1-3.414 0L8.61 19.61a1 1 0 0 0-1.68.474 2.5 2.5 0 1 1-3.014-3.015 1 1 0 0 0 .474-1.68l-1.683-1.682a2.414 2.414 0 0 1 0-3.414L4.39 8.61a1 1 0 0 1 1.68.474 2.5 2.5 0 1 0 3.014-3.015 1 1 0 0 1-.474-1.68l1.683-1.682a2.414 2.414 0 0 1 3.414 0z",key:"w46dr5"}]]};Lg.node;const Lx=ce(Lg);const Ug={name:"refresh-cw",size:24,node:[["path",{d:"M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8",key:"v9h5vc"}],["path",{d:"M21 3v5h-5",key:"1q7to0"}],["path",{d:"M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16",key:"3uifl3"}],["path",{d:"M8 16H3v5",key:"1cv678"}]]};Ug.node;const Ux=ce(Ug);const Hg={name:"rotate-ccw-clock",size:24,node:[["path",{d:"M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8",key:"1357e3"}],["path",{d:"M3 3v5h5",key:"1xhq8a"}],["path",{d:"M12 7v5l4 2",key:"1fdv2h"}]],aliases:["history"]};Hg.node;const Ig=ce(Hg);const Gg={name:"route",size:24,node:[["circle",{cx:"6",cy:"19",r:"3",key:"1kj8tv"}],["path",{d:"M9 19h8.5a3.5 3.5 0 0 0 0-7h-11a3.5 3.5 0 0 1 0-7H15",key:"1d8sl"}],["circle",{cx:"18",cy:"5",r:"3",key:"gq8acd"}]]};Gg.node;const Bg=ce(Gg);const Yg={name:"search",size:24,node:[["path",{d:"m21 21-4.34-4.34",key:"14j7rj"}],["circle",{cx:"11",cy:"11",r:"8",key:"4ej97u"}]]};Yg.node;const Hx=ce(Yg);const Xg={name:"send",size:24,node:[["path",{d:"M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z",key:"1ffxy3"}],["path",{d:"m21.854 2.147-10.94 10.939",key:"12cjpa"}]]};Xg.node;const Vg=ce(Xg);const Qg={name:"server",size:24,node:[["rect",{width:"20",height:"8",x:"2",y:"2",rx:"2",ry:"2",key:"ngkwjq"}],["rect",{width:"20",height:"8",x:"2",y:"14",rx:"2",ry:"2",key:"iecqi9"}],["line",{x1:"6",x2:"6.01",y1:"6",y2:"6",key:"16zg32"}],["line",{x1:"6",x2:"6.01",y1:"18",y2:"18",key:"nzw8ys"}]]};Qg.node;const Ix=ce(Qg);const Zg={name:"settings",size:24,node:[["path",{d:"M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915",key:"1i5ecw"}],["circle",{cx:"12",cy:"12",r:"3",key:"1v7zrd"}]]};Zg.node;const Gx=ce(Zg);const Kg={name:"shield-off",size:24,node:[["path",{d:"m2 2 20 20",key:"1ooewy"}],["path",{d:"M5 5a1 1 0 0 0-1 1v7c0 5 3.5 7.5 7.67 8.94a1 1 0 0 0 .67.01c2.35-.82 4.48-1.97 5.9-3.71",key:"1jlk70"}],["path",{d:"M9.309 3.652A12.252 12.252 0 0 0 11.24 2.28a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1v7a9.784 9.784 0 0 1-.08 1.264",key:"18rp1v"}]]};Kg.node;const Bx=ce(Kg);const Fg={name:"shrink",size:24,node:[["path",{d:"m15 15 6 6m-6-6v4.8m0-4.8h4.8",key:"17vawe"}],["path",{d:"M9 19.8V15m0 0H4.2M9 15l-6 6",key:"chjx8e"}],["path",{d:"M15 4.2V9m0 0h4.8M15 9l6-6",key:"lav6yq"}],["path",{d:"M9 4.2V9m0 0H4.2M9 9 3 3",key:"1pxi2q"}]]};Fg.node;const Jg=ce(Fg);const Wg={name:"sparkles",size:24,node:[["path",{d:"M11.017 2.814a1 1 0 0 1 1.966 0l1.051 5.558a2 2 0 0 0 1.594 1.594l5.558 1.051a1 1 0 0 1 0 1.966l-5.558 1.051a2 2 0 0 0-1.594 1.594l-1.051 5.558a1 1 0 0 1-1.966 0l-1.051-5.558a2 2 0 0 0-1.594-1.594l-5.558-1.051a1 1 0 0 1 0-1.966l5.558-1.051a2 2 0 0 0 1.594-1.594z",key:"1s2grr"}],["path",{d:"M20 2v4",key:"1rf3ol"}],["path",{d:"M22 4h-4",key:"gwowj6"}],["circle",{cx:"4",cy:"20",r:"2",key:"6kqj1y"}]],aliases:["stars"]};Wg.node;const Yx=ce(Wg);const Pg={name:"sun",size:24,node:[["circle",{cx:"12",cy:"12",r:"4",key:"4exip2"}],["path",{d:"M12 2v2",key:"tus03m"}],["path",{d:"M12 20v2",key:"1lh1kg"}],["path",{d:"m4.93 4.93 1.41 1.41",key:"149t6j"}],["path",{d:"m17.66 17.66 1.41 1.41",key:"ptbguv"}],["path",{d:"M2 12h2",key:"1t8f8n"}],["path",{d:"M20 12h2",key:"1q8mjw"}],["path",{d:"m6.34 17.66-1.41 1.41",key:"1m8zz5"}],["path",{d:"m19.07 4.93-1.41 1.41",key:"1shlcs"}]]};Pg.node;const Xx=ce(Pg);const $g={name:"terminal",size:24,node:[["path",{d:"M12 19h8",key:"baeox8"}],["path",{d:"m4 17 6-6-6-6",key:"1yngyt"}]]};$g.node;const Vx=ce($g);const ev={name:"thumbs-down",size:24,node:[["path",{d:"M9 18.12 10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79 1.11L12 22a3.13 3.13 0 0 1-3-3.88Z",key:"m61m77"}],["path",{d:"M17 14V2",key:"8ymqnk"}]]};ev.node;const tv=ce(ev);const nv={name:"thumbs-up",size:24,node:[["path",{d:"M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2a3.13 3.13 0 0 1 3 3.88Z",key:"emmmcr"}],["path",{d:"M7 10v12",key:"1qc93n"}]]};nv.node;const Hd=ce(nv);const av={name:"trash",size:24,node:[["path",{d:"M10 11v6",key:"nco0om"}],["path",{d:"M14 11v6",key:"outv1u"}],["path",{d:"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6",key:"miytrc"}],["path",{d:"M3 6h18",key:"d0wm0j"}],["path",{d:"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2",key:"e791ji"}]],aliases:["trash-2"]};av.node;const Dl=ce(av);const sv={name:"undo-2",size:24,node:[["path",{d:"M9 14 4 9l5-5",key:"102s5s"}],["path",{d:"M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5a5.5 5.5 0 0 1-5.5 5.5H11",key:"f3b9sd"}]]};sv.node;const Id=ce(sv);const iv={name:"users",size:24,node:[["path",{d:"M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2",key:"1yyitq"}],["path",{d:"M16 3.128a4 4 0 0 1 0 7.744",key:"16gr8j"}],["path",{d:"M22 21v-2a4 4 0 0 0-3-3.87",key:"kshegd"}],["circle",{cx:"9",cy:"7",r:"4",key:"nufk8"}]]};iv.node;const Qx=ce(iv);const ov={name:"volume-2",size:24,node:[["path",{d:"M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z",key:"uqj9uw"}],["path",{d:"M16 9a5 5 0 0 1 0 6",key:"1q6k2b"}],["path",{d:"M19.364 18.364a9 9 0 0 0 0-12.728",key:"ijwkga"}]]};ov.node;const Zx=ce(ov);const lv={name:"volume-x",size:24,node:[["path",{d:"M11 4.702a.7.7 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.7.7 0 0 0 11 19.298z",key:"1p7khw"}],["path",{d:"m16.5 14.5 5-5",key:"cul3yw"}],["path",{d:"m16.5 9.5 5 5",key:"1akey5"}]]};lv.node;const Kx=ce(lv);const rv={name:"x",size:24,node:[["path",{d:"M18 6 6 18",key:"1bl5f8"}],["path",{d:"m6 6 12 12",key:"d8bk6v"}]]};rv.node;const wa=ce(rv);async function Ji(o,c,u){const r=await fetch("/api/v1/auth/"+o,{method:"POST",credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1",...u?{Authorization:"Bearer "+u}:{}},body:JSON.stringify(c)}),h=await r.json();if(!r.ok)throw new Error(h.error?.message||"Authentication request failed.");return h}function Fx({onAuthenticated:o}){const[c,u]=f.useState(null),[r,h]=f.useState(""),[m,g]=f.useState(""),[v,b]=f.useState(""),[C,x]=f.useState(""),[y,w]=f.useState(!1),[A,E]=f.useState(!1),[_,N]=f.useState(""),[D,H]=f.useState(!1),G=f.useRef(null);f.useEffect(()=>{const S=new AbortController;return fetch("/api/v1/auth/status",{signal:S.signal,credentials:"same-origin"}).then(U=>{if(!U.ok)throw new Error("Cannot reach the local gateway.");return U.json()}).then(U=>{u(U),U.authenticated&&(!U.configured&&!U.disabled?E(!0):o())}).catch(U=>{U.name!=="AbortError"&&N(U.message)}),()=>S.abort()},[]),f.useEffect(()=>{y?G.current?.showModal():G.current?.close()},[y]);async function Q(S){S.preventDefault(),H(!0),N("");try{if(A){if(m!==v)throw new Error("Passwords do not match.");await Ji("settings",{username:r,password:m,disabled:!1}),g(""),b(""),o()}else c?.configured?(await Ji("login",{username:r,password:m}),g(""),o()):await Y()}catch(U){N(U.message)}finally{H(!1)}}async function Y(){await Ji("recover",{},C.trim()),x(""),w(!1),c?.configured?o():E(!0)}return a.jsxs("main",{className:"login",children:[a.jsxs("div",{className:"brand",children:[a.jsx("img",{className:"service-logo",src:Cd,alt:"",width:"104",height:"104"})," aiport",a.jsx("span",{className:"preview",children:"NORD"})]}),a.jsxs("section",{className:"login-card",children:[a.jsx("p",{className:"eyebrow",children:"YOUR LOCAL AI GATEWAY"}),a.jsx("h1",{children:A?"Make it yours.":c?.configured?"Welcome back.":"Start with your token."}),a.jsx("p",{className:"muted",children:A?"Choose a password for future sign-ins. A username is optional. Your startup token remains your recovery key.":c?.configured?"Sign in to your gateway. Your browser session stays active across reloads.":"For first-run authentication, use the management token from the terminal or private credential file that started aiport."}),_&&a.jsx("div",{className:"notice error",role:"alert",children:_}),a.jsxs("form",{onSubmit:Q,children:[A||c?.configured?a.jsxs(a.Fragment,{children:[a.jsxs("label",{htmlFor:"username",children:["Username",A?" (optional)":""]}),a.jsx("input",{id:"username",autoComplete:"username",value:r,onChange:S=>h(S.target.value),maxLength:64,required:!A&&c?.username_required}),a.jsx("label",{htmlFor:"password",children:A?"New password":"Password"}),a.jsx("input",{id:"password",type:"password",autoComplete:A?"new-password":"current-password",value:m,onChange:S=>g(S.target.value),minLength:A?12:void 0,maxLength:1024,required:!0}),A&&a.jsxs(a.Fragment,{children:[a.jsx("label",{htmlFor:"repeat",children:"Confirm password"}),a.jsx("input",{id:"repeat",type:"password",autoComplete:"new-password",value:v,onChange:S=>b(S.target.value),required:!0}),a.jsx("p",{className:"caption",children:"Use at least 12 characters. Only a salted password hash is stored."})]})]}):a.jsxs(a.Fragment,{children:[a.jsx("label",{htmlFor:"startup-token",children:"Startup management token"}),a.jsx("input",{id:"startup-token",type:"password",autoComplete:"off",value:C,onChange:S=>x(S.target.value),required:!0})]}),a.jsxs("button",{className:"primary",disabled:D||!c,children:[D?"Please wait…":A?"Save login and open gateway":c?.configured?"Sign in":"Verify token and set up login"," ",a.jsx(dn,{size:16,"aria-hidden":!0})]})]}),a.jsx("button",{className:"text-button forgot",onClick:()=>{N(""),w(!0)},children:"Forgot username or password?"}),a.jsx("p",{className:"caption",children:"Sessions expire after 12 hours or a server restart. You can disable local authentication in Settings."})]}),a.jsxs("p",{className:"login-footer",children:[a.jsx("span",{className:"dot"})," Nord · Local by default · No remote analytics"]}),a.jsxs("dialog",{ref:G,onClose:()=>{w(!1),x("")},className:"recovery-dialog",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{className:"recovery-title",children:[a.jsx("img",{className:"service-logo",src:Cd,alt:"",width:"28",height:"28"}),"Recover access"]}),a.jsxs("button",{className:"text-button",onClick:()=>w(!1),"aria-label":"Close recovery popup",children:["Close ",a.jsx(wa,{size:14,"aria-hidden":!0})]})]}),a.jsxs("div",{className:"recovery-content",children:[a.jsx("p",{children:"Your startup management token bypasses username/password login. After signing in, open Settings → Authentication to replace either credential."}),a.jsx("p",{children:"In the terminal that started aiport:"}),a.jsxs("pre",{children:[`printf '%s\\n' "$`,c?.token_env||"AIPORT_LOCAL_TOKEN",'"']}),a.jsx("p",{children:"Or, from the directory containing your onboarding files:"}),a.jsx("pre",{children:"./bin/aiport auth token --config aiport.yaml"}),a.jsx("p",{className:"caption",children:"Use the same config path as your server. This command reads the configured token environment variable, or the private aiport.yaml.env file. Agent and provider API keys cannot recover dashboard access. The recovery token is never shown by the web API."}),a.jsxs("form",{onSubmit:async S=>{S.preventDefault(),H(!0),N("");try{await Y()}catch(U){N(U.message)}finally{H(!1)}},children:[a.jsx("label",{htmlFor:"recovery-token",children:"Paste management token"}),a.jsx("input",{id:"recovery-token",type:"password",autoComplete:"off",value:C,onChange:S=>x(S.target.value),required:!0}),_&&a.jsx("p",{className:"auth-error",role:"alert",children:_}),a.jsxs("button",{className:"primary",disabled:D,children:["Bypass login and recover access ",a.jsx(dn,{size:16,"aria-hidden":!0})]})]})]})]})]})}function Jx({username:o,onChange:c}){const[u,r]=f.useState(o),[h,m]=f.useState(""),[g,v]=f.useState(""),[b,C]=f.useState(!1),[x,y]=f.useState(""),[w,A]=f.useState(""),[E,_]=f.useState(!1);f.useEffect(()=>{r(o),fetch("/api/v1/auth/status").then(D=>D.json()).then(D=>C(D.disabled)).catch(()=>y("Could not read authentication status."))},[o]);async function N(D){D.preventDefault(),_(!0),y(""),A("");try{if(h!==g)throw new Error("Passwords do not match.");await Ji("settings",{username:u,password:h,disabled:b}),m(""),v(""),A("Authentication settings saved. Previous browser sessions were revoked."),c()}catch(H){y(H.message)}finally{_(!1)}}return a.jsxs("section",{className:"panel auth-settings",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Authentication"}),a.jsx("span",{children:b?"Disabled":"Enabled"})]}),a.jsxs("form",{onSubmit:N,children:[a.jsxs("label",{className:"check-label",children:[a.jsx("input",{type:"checkbox",checked:b,onChange:D=>C(D.target.checked)}),"Disable local authentication"]}),a.jsx("p",{className:"caption",children:"When disabled, local programs and people with access to this machine can use the gateway and dashboard without credentials. Host and cross-site protections stay active. Recovery credentials are retained."}),a.jsxs("div",{className:"auth-fields",children:[a.jsxs("label",{children:["Username (optional)",a.jsx("input",{value:u,onChange:D=>r(D.target.value),autoComplete:"username",maxLength:64})]}),a.jsxs("label",{children:["New password",a.jsx("input",{type:"password",autoComplete:"new-password",value:h,onChange:D=>m(D.target.value),minLength:12,maxLength:1024,placeholder:"Leave blank to keep the current password"})]}),a.jsxs("label",{children:["Confirm new password",a.jsx("input",{type:"password",autoComplete:"new-password",value:g,onChange:D=>v(D.target.value)})]})]}),a.jsx("p",{className:"caption",children:"To reset the username, also enter a new password. Your management token can always bypass the login through the recovery popup."}),x&&a.jsx("p",{className:"auth-error",role:"alert",children:x}),w&&a.jsx("p",{className:"auth-success",role:"status",children:w}),a.jsx("button",{className:"primary",disabled:E,children:E?"Saving…":"Save authentication settings"})]})]})}const _d=`# aiport manual

aiport is a small gateway that runs on your own machine. Your apps and coding
agents send their LLM calls to aiport instead of straight to OpenAI, Anthropic
and the rest, and aiport sends each call on to the right provider.

That puts three things in one place. Your provider keys live in aiport, not in
every app. You decide which model serves what, and you can change it without
touching any app's code. And every call is recorded, so you can see what was
called, by whom, and how well it went.

This manual is written for you, the Owner: the person who runs aiport on their
machine and decides what it does. Other readers have their own documents.
Coding agents that connect an app read the
[communication standard](#the-communication-standard) and the briefs at
\`/_connect\`. The exact configuration keys are in
[configuration](configuration.md), the wire protocol is in
[Connect](connect.md), and policies and guardrails are in
[Extensions](extensions.md).

Words with a capital letter (Application, Connection, Level, Catalog, Issue,
Change and the rest) are aiport's own terms. [Core concepts](#core-concepts)
explains them, and so does the [glossary](../GLOSSARY.md).

## At a glance

Here is each of aiport's main ideas in one line.
[Core concepts](#core-concepts) explains them properly.

| Concept | In one line |
| --- | --- |
| Connection | A provider aiport sends calls to (OpenAI, Anthropic, Bedrock, Meta, Ollama, llama.cpp), with its key kept by aiport. |
| Application | Anything that calls aiport with its own token. Every call is recorded against it. |
| Protocol | \`openai\` or \`anthropic\`. aiport passes each call on in the provider's own format and never translates between them. |
| Level | \`fast\`, \`balanced\` or \`best\`: a model name an Application sends to say how capable a model it needs. |
| Task name | A model name for one kind of work (\`summarize\`, \`classify\`) that you point at a model. |
| Effort | \`low\`, \`medium\` or \`high\`, sent with \`X-Aiport-Effort\`: how hard one call is. |
| Module | The part of the Application a call serves, sent with \`X-Aiport-Module\`. Votes are read per module. |
| Catalog | What one Application can use through aiport, with a version on every answer. |
| Issue | A problem aiport keeps seeing for one Application, with evidence and a fix on each side. |
| Change | One configuration change made through aiport. Every Change can be reverted. |
| Suggestion | A change aiport proposes from request history or from votes. You decide. |
| Communication standard | The rules every integration follows, at \`/_connect?format=standard\`. |

## Quickstart

In about five minutes you will build aiport, connect one provider, and make
one real call through it, which you can then find in **Activity**.

All you need is Go 1.27 or newer. Nothing else gets downloaded, because the
dependencies are vendored and the dashboard comes pre-built.

1. Build aiport:

   \`\`\`sh
   git clone https://github.com/harelix/aiport && cd aiport
   make build                 # -> bin/aiport, dashboard included
   \`\`\`

   If you would rather not clone the repository, install the binary with
   \`go install github.com/harelix/aiport/cmd/aiport@latest\`. It lands in
   \`$(go env GOPATH)/bin\`.

2. Create the configuration and check it:

   \`\`\`sh
   ./bin/aiport init          # choose providers, coding agents and Levels
   ./bin/aiport check         # validate the configuration
   \`\`\`

   \`init\` writes \`aiport.yaml\` and, next to it, a private file for
   credentials, \`aiport.yaml.env\` (mode 0600). It never overwrites a file, and
   it never asks for provider keys in the terminal. It also proposes a first
   mapping for the Levels, as described in
   [Map Levels with init](#map-levels-with-init).

3. Start the server:

   \`\`\`sh
   ./bin/aiport serve         # prints the dashboard address and the management token
   \`\`\`

   When it runs in a terminal, \`serve\` prints the management token. If you
   need the token again later, \`./bin/aiport auth token\` shows it.

4. Open the address \`serve\` printed (\`http://127.0.0.1:8787/\` by default) and
   paste the management token to sign in. Then set a password of at least 12
   characters and, if you like, a username.

5. On **Providers**, paste the provider's API key into its card and click
   **Check connection**. The check makes one free \`GET /models\` call and
   generates nothing. A local provider such as Ollama or llama.cpp needs no
   key, but you have to start it yourself.

6. Map a Level. A Level is a shared model name such as \`fast\`, and mapping it
   tells aiport which model it means. On **Routing**, in the **Levels** panel,
   find \`fast\` on the OpenAI protocol. Click **Map** (or **Edit**, if \`init\`
   already mapped it) and choose a Connection and a model.

7. On **Integrate**, click **Add with a token**, name the Application
   \`quickstart\`, and copy the two \`.aiport\` lines it shows. Save them as a file
   named \`.aiport\` in an empty directory, and load them into your shell from
   there:

   \`\`\`sh
   set -a; . ./.aiport; set +a
   \`\`\`

8. Make the first call:

   \`\`\`sh
   curl -s http://127.0.0.1:8787/v1/chat/completions \\
     -H "Authorization: Bearer $AIPORT_APP_TOKEN" \\
     -H "X-Aiport-Module: quickstart" \\
     -H "X-Aiport-Effort: low" \\
     -H 'content-type: application/json' \\
     -d '{"model":"fast","messages":[{"role":"user","content":"Say hello"}]}'
   \`\`\`

   The answer comes from the model you mapped to \`fast\`. Open **Activity** and
   you will find the call listed with the name you sent (\`fast\`), its effort,
   and the Connection and model that served it.

To connect a real app next, hand its coding agent the communication standard,
as described in [Connect an application](#connect-an-application).
[Core concepts](#core-concepts) explains the words the rest of this manual
uses.

## Core concepts

The rest of the manual builds on these ideas. For each one you get what it is,
and then why it exists.

A **Connection** is a provider aiport can send calls to: OpenAI, Anthropic,
Amazon Bedrock, Meta Model API, or a local server such as Ollama or llama.cpp.
It holds an address and a key, a bit like an entry in your address book that
also keeps the key. Your provider keys live here, in one place, and never in
your apps.

An **Application** is anything that calls aiport with its own token: a script,
a service, a desktop app or a coding agent. It gets the token by pairing,
which you approve, much as you would let a new device join your Wi-Fi. Because
each app has its own token, every call is recorded against someone, and aiport
can tell each app about its own problems.

Each Connection speaks one **Protocol**, \`openai\` or \`anthropic\`. aiport passes
Chat Completions, Responses and Messages calls on in the provider's own format,
streaming included, and it never translates from one protocol to the other. So
a call from the OpenAI SDK needs a Connection that speaks \`openai\`, and a call
from the Anthropic SDK needs one that speaks \`anthropic\`. Since nothing is
translated, nothing gets lost in translation.

A **Level** is one of three model names that every aiport shares: \`fast\`,
\`balanced\` and \`best\`. An app sends a Level as its model name to say how
capable a model it needs, without naming a provider. It is like ordering
small, medium or large and letting the kitchen pick the cup. You map each
Level, per protocol, to a Connection and a model, and you can give one app its
own mapping. After that you can change models without anyone changing code.

A **task name** is a model name for one kind of work, such as \`summarize\`,
\`review-code\` or \`classify\`, that you point at a Connection and a model. Under
the hood it is an Alias, a name aiport swaps for a real model. Every app can
send a shared task name marked \`everywhere\`, and a task name with a fixed
\`reasoning\` effort always sends that effort to the provider. Task names let
routine work run on a cheaper model while the Levels stay on stronger ones.

**Effort** says how hard one call is: \`low\`, \`medium\` or \`high\`, sent in the
\`X-Aiport-Effort\` header. A Level or task name can have a different model for
each effort. If the call's body sets no effort of its own, the effort from the
header also reaches the provider as its reasoning effort. This way a call can
ask for more or less thinking without picking another model name.

A **Module** is the part of the app a call serves, such as \`tldr\`, \`classify\`
or \`insights\`, sent in the \`X-Aiport-Module\` header. aiport records it with the
call and never sends it on to the provider. Modules let you read the votes
people give answers one module at a time, and move one module to another model
while the others stay where they are.

The **Catalog** is the menu for one app: everything it can use through aiport.
That means its Levels, its task names and the models of the Connections it can
reach, plus a short version. The app reads it at \`GET /v1/models\`, and every
answer carries the version in the \`X-Aiport-Catalog\` header. The app never has
to guess a model name, and it learns at once when you change something.

A **Change** is one configuration change made through aiport, whether by you,
by a request you approved, or by an automatic fix you allowed. Every Change is
listed on the Changes page and can be reverted, like an undo history. That is
what makes it safe to experiment.

An **Issue** is a problem aiport keeps seeing for one app, such as a model that
no Connection serves. It comes with the evidence and two fixes: one in aiport,
which you approve, and one in the app, for its coding agent. Once calls
succeed again, it resolves itself. Issues make sure a failing integration gets
noticed and explained, instead of being retried forever.

A **Suggestion** is a change aiport proposes, and you always decide. Some come
from request history and appear on the Changes page, for example when an
unknown token keeps trying to call. Others come from the votes, the thumbs up
and down people give answers. Those appear on the Feedback page and are
applied as module routes.

A **Module route** is a route for one module of one app. Calls from that
module that send a given name go to another Connection and model, ahead of
every other route, while the app's other modules keep their routes. A module
that does badly can move without moving the whole app.

A **Change request** is an app asking for something aiport lacks, such as a
task name, along with what it needs and why. Only apps you allow may send one,
and you decide it on the Changes page.

The **Guidance version** is the version number of the guidance aiport gives
coding agents: the briefs, the \`AGENTS.md\` line and the MCP instructions. It
is now at version 4. aiport remembers the version each app's coding agent last
confirmed and tells the agent when it is behind, so integrations keep up as
aiport changes.

The **Communication standard** is one document, at
\`/_connect?format=standard\`, with every rule an integration follows and a
checklist: the token in \`.aiport\`, the SDK pointed at aiport, Levels and task
names, \`X-Aiport-Effort\`, \`X-Aiport-Module\`, the thumbs up and down, and the
fixes for errors. One paste brings any app in line, whether it is new or
already connected.

## Guides

Each guide takes one task from start to finish. For a page-by-page description
of the dashboard, see [A tour of the dashboard](#a-tour-of-the-dashboard).

### Connect an application

To connect an application, you give its coding agent one instruction and then
approve the code it shows you. The agent sends the app's LLM calls through
aiport and pairs the app. From then on the app has its own aiport token, sends
its calls to aiport, and never holds a provider key.

The quickest way takes three steps:

1. On **Integrate**, open the **Onboard an application** wizard and click
   **Copy instruction**.
2. Open the coding agent in the application's repository and paste the
   instruction.
3. When the agent shows you a pairing code, check that the same code appears
   on Integrate, then click **Approve**.

The rest of this section explains each part.

#### The communication standard

The communication standard is one document that lists every rule an
integration follows, with a checklist at the end. It is written for the coding
agent working in the application's repository, and it is the fastest way to
bring any application in line, new or already connected.

To hand it to a coding agent, go to **Integrate** and find the **Communication
standard** card at the top of **Instructions**. Click **Copy instruction** and
paste the prompt to the agent. **Copy URL only** copies just the address, and
**Read the standard** opens the document.

The standard has seven parts, and all of them are required.

1. Identity comes first. Each application has one token, kept in \`.aiport\` at
   the repository root and listed in \`.gitignore\`.
2. For transport, the OpenAI SDK points at \`AIPORT_BASE_URL + "/v1"\` and the
   Anthropic SDK at \`AIPORT_BASE_URL\`, both with \`api_key = AIPORT_APP_TOKEN\`.
   No provider keys stay in the application.
3. Model names are always a Level or a task name, never a provider's model id.
   \`GET /v1/models\` lists them.
4. The headers carry the rest. \`X-Aiport-Module\` goes on every model call and
   is required, and \`X-Aiport-Effort\` goes where calls differ in difficulty.
   The app keeps \`X-Aiport-Request-Id\` with every answer, together with
   \`X-Aiport-Model\` and \`X-Aiport-Route\`, which name the model that answered
   and how it was chosen. When \`X-Aiport-Catalog\` changes, it reads
   \`/v1/models\` again.
5. Rating means a thumbs up and a thumbs down next to every LLM answer the
   application shows a person. The coding agent builds them into the app's own
   UI, and they are sent to \`POST /v1/aiport/feedback\`.
6. For errors, the app follows \`error.fix\` on \`401\` and \`403\`. It retries with
   backoff only on \`429 concurrency_limit\`, \`502\` and \`504\`.
7. Talking to aiport is about the coding agent: it uses the MCP tools at the
   start of each session, files Change requests instead of building
   workarounds, and adds aiport's line to \`AGENTS.md\`.

An application without a user interface (a service, a batch job or a
command-line tool) offers the vote wherever a person does see the answer. If no
person ever sees it, the coding agent tells you so and skips the votes, but it
still sends \`X-Aiport-Module\`.

The document is served at \`/_connect?format=standard\`, and it is also the MCP
tool \`aiport_standard\`. The \`/_connect\` JSON gives it as \`standard\`, and the
prompt as \`standard_prompt\`.

#### Onboarding an application, step by step

The **Onboard an application** wizard on Integrate walks you through connecting
one application. Each step ticks itself off as aiport sees it happen, and the
step to do now is highlighted.

The wizard is open while you have no Applications and whenever a pairing
request is waiting. Otherwise it is collapsed, and its header tells you where
things stand, such as "1 pairing waiting", or the Application it is following
and how far it got. To follow an app, select an existing Application in the
list, or click **Onboard a new application** for a new one.

1. Open the coding agent (Claude Code, Codex, Cursor and so on) in the
   application's repository. Under **Hand this to the coding agent**, click
   **Copy instruction** and paste it to the agent (**Copy URL only** copies
   just the address). [What the instruction asks](#what-the-instruction-asks)
   below explains what it says.
2. Next comes **Approve the pairing**. The agent asks aiport to pair, and
   within seconds the request appears at the top of Applications with a short
   code. Check that it matches the code the agent shows you, then click
   **Approve**. While you approve, you can also bind a Connection and allow
   change requests. Requests wait 30 minutes. Once you approve, the wizard
   follows that Application.
3. The agent makes one real call. The step **First call arrives** ticks when
   Activity has a call from the Application in the last 7 days.
4. You can switch on **Allow change requests** with the toggle in that step,
   which is optional. You still decide each request yourself in Changes.
5. The optional **MCP entry in the repository** step is usually done already,
   because the agent adds the entry in step 1. Tick it by hand, or copy the
   entry from the **Coding agent brief** tile.
6. Finally, the optional **Connect a coding agent** step ticks when the
   Application has a coding agent (its own or the default) and a repository
   path. See [Connecting a coding agent](#connecting-a-coding-agent).

##### What the instruction asks

The instruction tells the agent to read
\`http://127.0.0.1:8787/_connect?format=markdown\` and to do two parts, in
order:

1. Wire the LLM calls through aiport. That means pairing, saving the token in
   \`.aiport\` and adding that file to \`.gitignore\`, reading \`.aiport\` at
   startup, pointing the SDK at aiport, removing provider keys, sending
   \`X-Aiport-Module\`, adding the thumbs up and down, and making one test call.
2. Add the aiport MCP entry and the \`AGENTS.md\` line.

It also points the agent at the communication standard for the checklist. The
agent is asked to tell you the pairing code and wait, never to commit
\`.aiport\`, and to ask you for a **New token** instead of pairing again if the
token is lost.

The markdown document opens with a **Start here** section and holds everything
for both parts, so an agent given only that address needs nothing else. The
exact instruction text, with your gateway's address, is at the top of that
document, and the \`/_connect\` JSON has it as \`agent_prompt\`.

#### Applications and pairing

An Application is anything that calls aiport with its own token, and pairing is
how it gets that token with your approval. aiport credits each call to the
Application that sent it, so Activity shows who called what. Calls without an
Application token show as \`unknown\`.

Pairing goes like this:

1. The application asks aiport to pair, giving its name, and shows you a short
   code. It may also report the path of its git repository on this machine,
   which the Application brief asks the integrating coding agent to do.
2. The request appears with that code at the top of the Applications list on
   **Integrate**, and in the Applications table on **Routing**. Check that the
   application shows the same code.
3. If you want, choose a Connection to bind it to and tick **May request
   changes**.
4. Click **Approve**, or **Reject**.
5. The application collects its token, once. aiport adds the Application to
   the configuration without a restart and stores the token in the private
   credential file as \`AIPORT_APP_<ID>_TOKEN\`.

##### The repository path

If the application reported a repository path, it appears in a field under the
request. If the path is not a git repository on this machine, the field says
what is wrong with it. Confirm, edit or clear the path before you approve.

aiport keeps only what the field holds when you approve, and it never uses a
reported path you did not confirm. A path that does not exist, or is not the
top directory of a git repository, is refused with the reason, and the request
keeps waiting.

##### Where the token lives: \`.aiport\`

Each application keeps its aiport settings in a file named \`.aiport\` at the
root of its repository. The file is listed in \`.gitignore\` and never
committed:

\`\`\`
AIPORT_BASE_URL=http://127.0.0.1:8787
AIPORT_APP_TOKEN=<its token>
\`\`\`

The application loads these into its environment at startup, and variables
that are already set win.

Its coding agents read the file too. The project MCP entry's \`headersHelper\`
takes the token from \`AIPORT_APP_TOKEN\`, or else from \`.aiport\`, so nothing
has to be exported before the agent starts.

If a token is lost or revoked, aiport answers \`401\`. To fix it, click **New
token** on the Application's row on Integrate, then click **Copy .aiport
lines** and paste them over the app's \`.aiport\`. The old token stops working
at once.

Applications integrated before the token moved to \`.aiport\` were told to move
it there by guidance version 3.

##### One request per application

If an application asks again while its request is still waiting, it gets the
same request back, with the same code, so the code you see stays valid. aiport
treats it as the same application when it has the same repository, or the same
name ignoring case and punctuation. Its earlier and newer poll paths both keep
working.

A request that matches an Application you already have is marked **Already
connected as …**. Its main button, **Give … a new token**, approves it as that
existing Application, which receives a new token while its old one stops
working. This avoids a duplicate such as \`docs-bot-2\`. **Add as separate**
still adds a new Application.

##### Waiting and restarts

Requests wait 30 minutes for you, and they survive an aiport restart because
the poll secret is stored only as a hash. After a restart you can still
approve a waiting request by its code, and an approved application can still
collect its token for 24 hours. Once a request has run out, the application's
poll answers \`expired\`, so it knows to ask again.

##### Rejected requests

When you **Reject** a request, aiport remembers it and the application learns
it was rejected. For the newest 100 rejections, aiport keeps the name,
description, code and time (never a token), across restarts. The **Rejected**
filter on Integrate shows them, and **Forget** removes one.

A rejected application may ask to pair again.

##### The Applications table

The **Applications** table on Routing lists every Application with its client
id, its Connection, its own Aliases (model names that only it uses), and the
**May request changes** toggle.

Over the API, the application starts pairing with \`POST /v1/aiport/pair\` and
its name. No token is needed, and the route is public on loopback and
rate-limited. A poll answers \`pending\`, \`rejected\`, \`expired\`, or \`approved\`
with the token. To approve a request as an existing Application over the API,
send \`{"replace": "<id>"}\`.

#### Give a token yourself, or remove an Application

You can skip pairing and give an application its token yourself. This helps
when its coding agent is not running, or when you would rather paste the token
into its \`.aiport\` file.

To add one, click **Add with a token** on Integrate, next to **Onboard a new
application**. Name the Application and, if you want, choose a Connection.
aiport adds it and shows its token once, on a token card with **Copy** and
**Copy .aiport lines**. The lines are \`AIPORT_BASE_URL\` (this gateway's
numeric address) and \`AIPORT_APP_TOKEN\`.

To replace a token, use **New token**, the key button at the end of an
Application's row (or **Issue new token** after you click its name). Once you
confirm, aiport issues a new token and shows it once, the same way. The old
token stops working at once, whether it came from pairing or from here.
Issuing a token changes no configuration, so it does not appear in Changes.

To remove an Application, use **Remove**, the trash button at the end of the
row (or **Remove application** after you click its name). aiport first lists
what goes with it: its own Aliases, its binding, its coding agent settings and
its permission to request changes. It then removes all of it as one Change and
deletes the token, which stops working at once. The Application's activity
history is kept.

Two cases need a little care. If you revert a removal in **Changes**, the
Application comes back **without a token**. Its row shows **No token**, and
its calls are refused until you click **Issue token**. \`serve\` and \`aiport
check\` also name such Applications. The other case is a token that was set in
aiport's own environment when it started (exported in the shell instead of
kept in the private credential file). The dashboard cannot replace or delete
it, so change it where aiport is started, and restart aiport.

Scripts can do all three with the management token: \`POST /api/v1/origins\`
with \`{name, target?}\`, \`POST /api/v1/origins/{id}/token\` and \`DELETE
/api/v1/origins/{id}\`. Each needs \`X-Aiport-CSRF: 1\`. A token appears in the
response once and is never shown again.

### Set up the application's coding agent

The coding agent that works in an application's repository talks to aiport
over MCP, using the Application's own token. That lets it read the Catalog,
its Issues and any new guidance at the start of every session.

You set this up once per repository, with an MCP entry and one line in
\`AGENTS.md\`. If you onboarded the app with the wizard, the integrating agent
already added both in step 1.

To add them yourself:

1. Open the **Coding agent brief** on Integrate. It has both pieces, ready to
   copy.
2. Add the MCP entry to the repository's \`.mcp.json\`.
3. Add the line to the repository's \`AGENTS.md\`.

#### The MCP entry and the AGENTS.md line

The project MCP entry, in \`.mcp.json\`, points at this gateway's \`/mcp\` endpoint
over HTTP. Its \`headersHelper\` reads the token from \`AIPORT_APP_TOKEN\`, or from
\`.aiport\` at the repository root when that variable is unset, so the entry
itself never holds the token.

The \`AGENTS.md\` line tells every later session to:

- find the token in \`.aiport\`, and what to do on \`401\`
- send a Level or task name, never a provider's model id
- state effort with \`X-Aiport-Effort\`
- name the module of every call with \`X-Aiport-Module\`
- add the thumbs up and down where the app has a UI
- check \`aiport_catalog\`, \`aiport_my_issues\` and \`aiport_guidance\` at the
  start of a session, and call \`aiport_guidance_done\` after updating
- ask with \`aiport_request_change\` instead of working around aiport.

When an agent connects over MCP (\`initialize\`), the instructions aiport sends
it say the same as the \`AGENTS.md\` line.

The MCP endpoint serves one Application at a time. A connection without a
recognized Application token is refused with \`403
application_token_required\`.

Claude Code reads the MCP token when it connects. After the token in \`.aiport\`
changes, run \`/mcp\` and reconnect aiport, or start a new session.

##### The MCP tools

Each tool answers exactly as its HTTP endpoint does, with the Application's own
token.

| Tool | HTTP equivalent | What it does |
| --- | --- | --- |
| \`aiport_standard\` | \`GET /_connect?format=standard\` | The communication standard, with its checklist. |
| \`aiport_capabilities\` | \`GET /v1/aiport/capabilities\` | Protocols, Connections, own Aliases, binding, \`may_request_changes\`, \`change_kinds\`, \`open_issues\`, \`catalog\` and \`guidance\`. |
| \`aiport_catalog\` | \`GET /v1/aiport/catalog\` | The Catalog and its recent changes, newest first. |
| \`aiport_guidance\` | \`GET /v1/aiport/guidance\` | What changed in the guidance since the confirmed version, the current \`AGENTS.md\` line and the steps. |
| \`aiport_guidance_done\` | \`POST /v1/aiport/guidance/done\` | Confirms the current guidance version. |
| \`aiport_request_change\` | \`POST /v1/aiport/requests\` | Files a Change request. |
| \`aiport_my_requests\` | \`GET /v1/aiport/requests\` (or \`/{id}\`) | The Application's Change requests and your answers. |
| \`aiport_my_issues\` | \`GET /v1/aiport/issues\` (or \`/{id}\`) | The Application's Issues. Offered only while adapting is on. |
| \`aiport_issue_fixed\` | \`POST /v1/aiport/issues/{id}/fixed\` | Says an Issue is fixed in the Application's code. Offered only while adapting is on. |

The Coding agent brief itself is served at
\`/_connect?format=handover&for=agent\`.

#### Guidance versions

aiport gives coding agents guidance: the briefs, the \`AGENTS.md\` line and the
MCP instructions. It numbers that guidance and raises the number whenever it
changes in a way agents must act on, and each version has a short change log.

aiport remembers, per Application, the last version its coding agent
confirmed, so an agent that is behind learns exactly what to do.

| Version | What it asks agents to do |
| --- | --- |
| 1 | The project setup: the MCP entry, the \`AGENTS.md\` line, and reading \`aiport_catalog\` and \`aiport_my_issues\` at the start of a session. |
| 2 | Levels and effort: send a Level or task name instead of a provider's model id, add \`X-Aiport-Effort\` where calls differ in difficulty, replace the \`AGENTS.md\` line, and check \`aiport_guidance\` at the start of each session. |
| 3 | The token lives in \`.aiport\`: add it to \`.gitignore\`, move the token there, load it at startup, replace the MCP entry with the one whose \`headersHelper\` reads \`.aiport\`, replace the \`AGENTS.md\` line, and ask for a **New token** on \`401\` instead of pairing again. |
| 4 | Votes per module: read the communication standard, give each LLM feature a module name sent as \`X-Aiport-Module\`, keep \`X-Aiport-Request-Id\` with each answer, build a thumbs up and down into the UI, and replace the \`AGENTS.md\` line. |

Every version ends with the agent calling \`aiport_guidance_done\`. The confirmed
version is kept in the database next to the configuration. Applications
integrated before guidance versions existed have none.

##### How a coding agent that is behind finds out

You do not have to tell it. While an Application is behind, aiport reminds its
coding agent in three places.

- The MCP \`initialize\` instructions start with a one-line reminder.
- Every aiport MCP tool answer ends with the same reminder, as a second text
  block. The exceptions are \`aiport_guidance\` and \`aiport_guidance_done\`,
  whose answers are the guidance itself.
- \`aiport_capabilities\` shows \`guidance: {current, confirmed}\`, where
  \`confirmed\` is \`null\` if the agent never confirmed.

The agent then calls \`aiport_guidance\`, does the steps, and calls
\`aiport_guidance_done\`. Both use the Application's own token and only show or
change that Application's state. All of this works with adapting off.

##### Where you see it

The dashboard shows where each Application stands in two places.

On Integrate, the **Guidance** column shows each Application's state. One that
is behind shows **Behind v… → v…**, or **Unknown** if it never confirmed. The
copy button next to it copies one paragraph to paste to its coding agent.

On **Overview**, one quiet line appears while any Application has not
confirmed the current version, and an Application that never confirmed any
version counts as not confirmed. The line tells the two cases apart: "behind"
means it confirmed an older version, and "never confirmed" means none yet.

Over the API, \`GET /v1/aiport/guidance\` (the \`aiport_guidance\` tool) returns
\`{current, confirmed, confirmed_at_ms, behind, changes, agents_line, steps}\`,
and \`POST /v1/aiport/guidance/done\` confirms. On the management side, \`GET
/api/v1/guidance\` has \`behind\` and \`never_confirmed\`.

### Levels

Levels are three model names every aiport shares: \`fast\`, \`balanced\` and
\`best\`. You decide which model each one means on your machine, per protocol.
An Application sends a Level as its model to say how capable a model it needs
without naming a provider, so you can move it to another model without
changing its code.

To map a Level:

1. Go to **Routing → Levels**.
2. Click **Edit** on the Level and protocol (or **Map**, if it is not mapped
   yet).
3. Choose a Connection and a model, then click **Save**.

\`aiport init\` and \`aiport.yaml\` can map Levels too. Both are covered below.

#### How a Level mapping works

A Level is a reserved Alias name, and an Alias is a model name that aiport
replaces with a real model on a Connection. A Level can be mapped in two ways.

The shared mapping is an entry under \`aliases:\` whose \`alias\` is the Level, on
a Connection of the protocol it serves. Each Level is mapped at most once per
protocol. An Application's own mapping is the same kind of entry under its
\`origins[].aliases\`, and it comes before the shared one, on that protocol
only.

\`\`\`yaml
aliases:
  - {target: hosted, alias: best, model: big-model}         # OpenAI protocol
  - {target: local, alias: fast, model: 'llama3:8b'}        # OpenAI protocol
  - {target: team, alias: balanced, model: team-model}      # Anthropic protocol
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: best, target: local, model: 'llama3:70b'}   # its own best, OpenAI only
\`\`\`

A Level mapping chooses the Connection. That is what sets it apart from a
plain shared Alias, which only renames a model on the Connection the request
is already going to. It chooses the Connection even for an Application bound
to another one, or for one with no protocol default.

Only the request's \`model\` field is replaced, except that in one case the
reasoning effort is added too (see
[Effort and reasoning](#effort-and-reasoning)).

Mapped Levels are listed in each Application's \`GET /v1/models\`, on the
protocol they are mapped on, and in its Catalog. Changing a mapping is a
Change you can revert, and also a Catalog change the Applications are told
about. If you remove a Connection that a Level is mapped on, aiport asks you
for a replacement Connection of the same protocol.

#### When a Level is not mapped

aiport never quietly uses another model. A Level that is not mapped on the
request's protocol is refused with \`400 level_unmapped\`, naming the Level and
the protocol. If an Application keeps asking for an unmapped Level, aiport
opens an Issue for it (see [An unmapped Level](#an-unmapped-level)).

Some requests have nowhere at all to go, and aiport refuses those at once,
before reading the body. That happens when, for the request's protocol, the
Application has none of these:

- a binding or a protocol default
- a task name of its own there
- a Level mapped for it
- an \`everywhere\` task name
- a module route.

Once any of those exists for the Application, or while its Connection is only
switched off, aiport reads the model first. A Level is then still refused as
\`level_unmapped\`, with its name.

#### Map Levels with init

\`aiport init\` suggests a first mapping per protocol, based on the providers
you chose. \`fast\` goes to a local Connection if you chose one, and \`balanced\`
and \`best\` go to hosted ones. If you chose only one kind, it uses whatever you
chose.

The presets do not know any model ids, so \`init\` asks you for each Level's
model. An empty answer leaves that Level unmapped, and nothing is mapped
unless you confirm.

Without a terminal, pass the mapping as a flag:

\`\`\`
--levels fast=local-connection/llama3:8b,best=hosted-connection/some-model
\`\`\`

Each item is \`LEVEL=CONNECTION/MODEL\`. The Connection is one of the provider
ids you chose, and the protocol is that Connection's. To skip Levels, pass
\`--no-levels\`. With neither flag and no input, no Level is mapped, and \`init\`
says so.

#### Map Levels in the dashboard

You can edit Levels in the dashboard without touching the file. Every save is
one Change, so you can revert it in Changes to get the previous mapping back.

**Routing → Levels** shows each Level on each protocol. **Edit** (or **Map**,
when the Level is not mapped) opens its mapping: a Connection of that
protocol, a model chosen from the models the Connection lists, and optional
per-effort models. **Clear** removes the shared mapping. Each cell also says
how many Applications have their own mapping.

For a single Application, use **Integrate → Applications → Levels**, the toggle
under its name next to **Catalog**. It shows the same grid for that
Application: its own mappings, marked **Own**, and the shared ones it falls
back to. **Override** gives it its own mapping, and **Use the shared mapping**
removes it again.

To check the result, the [Routing map](#the-routing-map) shows what each Level
reaches for each Application. Activity shows which name and effort each call
asked for, and what served it.

#### Upgrading: Aliases named like a Level

If your configuration already had an Alias called \`fast\`, \`balanced\` or
\`best\`, it is now that Level's mapping on its Connection's protocol.

A shared one now chooses the Connection for every Application's calls for that
name on that protocol, whatever they are bound to. Before, it only renamed a
model on the Connection a call was already going to. An Application's own one
applies on its Connection's protocol only.

\`aiport check\` prints one notice per such entry, naming it and its Connection,
and the startup log sums them up in one line. To keep the entry as the Level,
do nothing. To keep it as a plain Alias, rename it (its \`alias:\`) to a name
that is not a Level, and send the new name.

### Create task names for routine work

A task name is a model name for one kind of work, such as \`classify\`, \`tldr\`
or \`extract\`, that you point at a Connection and a model. A typical use is a
cheaper model for routine work at a fixed effort, while the Levels stay on the
models for harder work.

For example, to send all classification work, from every Application, to a
small local model with low reasoning effort, add this to \`aiport.yaml\`:

\`\`\`yaml
aliases:
  - {target: local, alias: classify, model: 'llama3:8b', everywhere: true, reasoning: low}
\`\`\`

Any Application can now send \`classify\` as its model.

#### Aliases and task names

An Alias is a model name an Application sends that aiport replaces with a real
model on a Connection. A task name is simply an Alias named for the work. Only
the \`model\` field of the request is rewritten.

Shared Aliases apply to every Application. They live under \`aliases:\` in
\`aiport.yaml\`, one per Connection and name, and without \`everywhere\` a shared
Alias applies only to calls that already go to its Connection. An
Application's own Aliases live under \`origins[].aliases\`. They come first, and
they may pick a different Connection from the Application's usual one.

For example, docs-bot has its own Alias \`summarize\` and the shared Alias
\`review-code\`, both on the local Connection it is bound to. That Connection
also serves \`llama3:8b\`. So docs-bot can send \`summarize\`, \`review-code\` or
\`llama3:8b\`.

You can write Aliases in the configuration file, or let them arrive as
Changes. Approving a Change request of kind \`application_alias\`, or an Issue's
fix, adds one. It is for that Application, or a shared Alias if you choose
**every Application** as the Scope.

An Application lists its own Aliases with \`GET /v1/aiport/capabilities\`, and
every name it can use with \`GET /v1/models\` (see [The Catalog](#the-catalog)).

#### Task names for every Application, with a fixed reasoning effort

Two settings on a shared task name make it work like a Level.

With \`everywhere: true\`, every Application can send it, whichever Connection
its calls otherwise go to, and the name routes to its own Connection. With
\`reasoning: low\`, \`medium\` or \`high\`, aiport sends that reasoning effort to
the provider on every call for the name, whatever \`X-Aiport-Effort\` says. A
call whose body already sets its own reasoning effort keeps it.

The example at the top of this page uses both. Each Application's Catalog
lists the name with its \`reasoning\`.

A few limits apply. A task name is 1 to 256 characters, with no spaces or
slashes, and it cannot be a Level (map those under [Levels](#levels)). One
name can route \`everywhere\` from only one Connection, so to move it, remove it
there first.

Scripts add or change a task name with \`PUT /api/v1/task-names/{name}\` and
\`{connection, model, effort?, everywhere?, reasoning?}\`, and remove it with
\`DELETE /api/v1/task-names/{name}\` and \`{connection}\`. Each is a Change you
can revert.

#### Per-effort models of a task name

A task name can use a different model for each effort (low, medium, high), all
on its own Connection.

1. Find the task name's row on the [Routing map](#the-routing-map), or in an
   Application's Catalog on Integrate.
2. Click the pencil.
3. Fill in a model for low, medium and high. An empty field uses the task
   name's usual model, and **Clear** removes them all.

A shared task name's per-effort models apply to every Application. Each save
is one Change you can revert.

From the API, use \`PUT /api/v1/aliases/effort\` with \`{application?, alias,
connection?, effort}\`. A shared task name is found by its name and
\`connection\`, an Application's own by its name. In \`aiport.yaml\` the
per-effort models are the Alias's \`effort:\` field.

### The Catalog

The Catalog is the list of names an Application, meaning an app or agent connected to aiport, can use through aiport. You decide what is in it and aiport tells the app, so the app never has to guess a model name or find out from a provider's \`404\` that something changed.

An Application's Catalog holds the Levels mapped for it, its task names, the
models its Connections list, and a short version code that changes whenever the
Catalog does. The task names come in a fixed order: the app's own Aliases first,
then the shared Aliases of the Connections it reaches, then the \`everywhere\`
names.

The briefs and the communication standard teach every app and its coding agent
how to use it. An app sends a Level or a task name, never a provider's model id.
It reads \`GET /v1/models\` to find the names, watches the \`X-Aiport-Catalog\`
header, and checks \`aiport_catalog\`, \`aiport_my_issues\` and \`aiport_guidance\` at
the start of each session.

The Catalog works with adapting turned off, because it is information, not a
Change. Only the Catalog-change Issues follow the adapting switch.

#### What an Application reads

To see an app's Catalog yourself, open Integrate and click **Catalog** on the
app's row. You get the same view the app gets, along with its recent Catalog
changes.

An app reads its Catalog with its own token, and every answer is about that app
alone. It never sees another app's Catalog. Keys, addresses, prompts and tokens
never appear.

Scripts, SDKs and coding agents reach the same information over HTTP, through
three answers and one header.

\`GET /v1/models\` is where OpenAI and Anthropic SDKs and coding agents already
look for models. It answers in the OpenAI list shape, or in the Anthropic shape
when the request carries \`anthropic-version\` (Anthropic SDKs may send the token
in \`X-Api-Key\`). Only names usable on that protocol are listed. Mapped Levels
and task names come first, marked as aiport's with \`owned_by: "aiport"\`, or with
"(aiport Level)" and "(aiport task name)" in the display name. After them come
the models of the Connection the app is bound to, or of the protocol defaults.
These are the same free, cached lists **Show models** uses, and a Connection
whose model list cannot be read adds nothing. Your management token gets
\`403 application_token_required\`, since a Catalog belongs to one Application.

\`X-Aiport-Catalog: <version>\` (12 hex characters) is on every answer to a
request made with an Application token, forwarded calls and errors included.
Apart from that header, the provider's answer is unchanged. The version changes
only when the app's task names or where they go, its binding, or the protocol
defaults it follows change. A provider adding or retiring a model does not
change it. An app that sees a new version reads \`/v1/models\` again.

\`GET /v1/aiport/catalog\` returns \`catalog\` and \`changes\`, and the MCP tool
\`aiport_catalog\` gives the same answer. \`catalog\` holds the task names with
their Connection and model. Levels are among them, with \`level: true\` and their
\`protocol\`. Per-effort models appear as \`effort\`, along with a fixed \`reasoning\`
where one is set. It also holds the binding or the protocol defaults followed,
the Connections reached, their models, and the version. \`changes\` holds the
app's recent Catalog changes, newest first. Here too, your management token gets
\`403 application_token_required\`.

The capabilities answer (\`GET /v1/aiport/capabilities\`) carries
\`catalog: {version, changed_at_ms}\`, which is the version and the time of the
newest kept Catalog change (\`null\` when none is kept).

#### The change log

Every configuration Change goes through one place. On each one, aiport works out
what it changed in every app's Catalog.

These are the Changes that can do that:

- approving or reverting an Alias
- mapping a Level, or changing its per-effort models
- adding a task name
- binding an Application
- changing a protocol default
- removing a Connection

A revert is a Catalog change too, with \`revert_of\` naming the Change it undid.
An app the Change did not affect gets nothing.

aiport's own model calls get Catalog changes recorded as well. They belong to
the built-in Application \`aiport\`, which follows \`agent_runs.connection\`, or the
protocol defaults when that is not set.

Catalog changes are kept as long as request history
(\`telemetry.retention_days\`) and survive a restart. They hold Application ids,
task names, model names and Connection ids, and nothing else: never keys,
addresses, prompts, responses or tokens.

Over the API, each Catalog change records:

- the Change id (\`change_id\`) and time (\`at_ms\`)
- the version \`before\` and \`after\`
- task names \`removed\`, \`added\` or \`moved\` (with \`from\` and \`to\`)
- the binding or protocol defaults that changed (\`routing\`)
- the Connections the app no longer reaches or newly reaches
  (\`connections_removed\`, \`connections_added\`)

#### When a Change takes something away

Sometimes a Change removes something an app was using. It might remove a task
name the app used, or move its binding or a protocol default it followed.

In that case aiport does more than change the version. It opens a Catalog-change
Issue (cause \`catalog_changed\`) for that app, saying what changed and what the
app can still use.

"Used" means that, within request history, the app's calls sent that task name,
or sent a model name that went through that binding or default. An app that
never did gets only the Catalog change.

The Issue needs no decision from you and resolves itself (see [A Catalog
change](#a-catalog-change)). Before you remove a Connection, the removal dialog
shows which apps it will affect, as described in [Removing a Connection, and
when one fails](#removing-a-connection-and-when-one-fails).

### Effort and reasoning

Effort says how hard one call is: \`low\`, \`medium\` or \`high\`. It lets aiport pick
a stronger or a cheaper model for the same name. For example, \`balanced\` with
high effort can go to a stronger model than \`balanced\` with low effort.

#### Where a call's effort comes from

aiport looks in three places, in this order, and takes the first it finds.

1. The \`X-Aiport-Effort\` header, set to \`low\`, \`medium\` or \`high\`. This is the
   one the briefs teach. It also accepts the provider words below, with \`none\`
   and \`minimal\` counting as low and \`xhigh\` and \`max\` as high. A value that is
   not an effort counts as no header.
2. The provider's own effort field, which is \`reasoning_effort\` on OpenAI Chat
   Completions, \`reasoning.effort\` on OpenAI Responses and
   \`output_config.effort\` on Anthropic Messages. Here \`none\`, \`minimal\` and
   \`low\` count as low, \`medium\` as medium, and \`high\`, \`xhigh\` and \`max\` as
   high. Numeric thinking budgets (\`budget_tokens\`) are never read or
   translated.
3. Otherwise, \`medium\`.

#### Per-effort models

An Alias is a name aiport turns into a model on a Connection, and any Alias can
have its own model for each effort, on its own Connection. That covers Levels
and task names, shared or an app's own. An effort with no model of its own uses
the Alias's \`model\`.

\`\`\`yaml
aliases:
  - {target: hosted, alias: balanced, model: mid-model, effort: {low: small-model, high: big-model}}
origins:
  - id: docs-bot
    # ...
    aliases:
      - {alias: summarize, target: local, model: 'llama3:8b', effort: {high: 'llama3:70b'}}
\`\`\`

Here \`balanced\` goes to \`small-model\` at low effort, \`mid-model\` at medium and
\`big-model\` at high. docs-bot's \`summarize\` goes to \`llama3:70b\` at high effort
and to \`llama3:8b\` otherwise.

You edit them in two places. A Level's per-effort models are in its mapping on
Routing or Integrate, where empty fields use the mapping's model. A task name's
are on its row of the Routing map.

Per-effort models are listed in the Catalog (\`effort\` on a task name), so
changing them is a Catalog change. Activity shows the effort of each call and
where it came from (see [Activity and history](#activity-and-history)).

#### What reaches the provider

The \`X-Aiport-Effort\` header itself is never forwarded, and neither is any other
\`X-Aiport-*\` request header. If the body already has the provider's effort
field, aiport never changes it.

If the body has no effort field and aiport chose the model, aiport adds one, so
the provider reasons that hard too. aiport counts as having chosen the model
when a Level, a task name, an Alias or a module route replaced the name sent.
What it adds depends on the call:

- A task name with a fixed \`reasoning\` gets that effort, whatever the header
  says.
- Otherwise, if the effort came from the \`X-Aiport-Effort\` header, aiport adds
  \`low\`, \`medium\` or \`high\`.
- Otherwise it adds nothing.

The field is \`reasoning_effort\` on Chat Completions, \`reasoning.effort\` on
Responses and \`output_config.effort\` on Messages. The rest of the body is kept
byte for byte. A model name sent unchanged, with no Alias applying to it, gets
nothing added.

Per-effort models can also be dropped. Over the API, a \`PUT\` on a Level without
\`effort\` keeps its per-effort models as long as the mapping stays on the same
Connection. If the mapping moves to another Connection, they are dropped,
because they name models of the old one. The same happens when a removal plan
moves a Level or a task name to another Connection: it keeps its model and
loses its per-effort models, so you set them again on the new Connection.

### Collect feedback and improve routing

People rate the answers an app shows them, and aiport turns those votes into a
report per module and into Suggestions you can apply. This shows you which
model and effort suit each part of each app, so you can move the part that does
badly without moving the rest.

#### Votes from the application

A module is one part of an app, such as its summarizer or its search. For
feedback to work, every integration does two things:

1. It names the module of each model call with the \`X-Aiport-Module\` header.
2. It gives every LLM answer shown to a person a thumbs up and a thumbs down.

The communication standard and guidance version 4 ask every connected app's
coding agent to add both.

A module name is lower case, 1 to 64 characters of \`a\`–\`z\`, \`0\`–\`9\`, \`.\`, \`_\`
and \`-\`, and does not start with a punctuation mark. It is recorded with the
call and never forwarded.

Votes need request history (\`telemetry.enabled\`) and count only calls still
retained. On Integrate, the **Guidance** column shows two chips for each
Application with calls in the last 7 days. **Modules** means its calls name
their module, and **Votes** means its users' votes arrive.

A vote carries only the request id, the vote and the module. Prompts, answers
and anything a person typed are never sent with it.

When a person votes, the app sends:

\`\`\`sh
curl -s -X POST http://127.0.0.1:8787/v1/aiport/feedback \\
  -H "Authorization: Bearer $AIPORT_APP_TOKEN" -H 'content-type: application/json' \\
  -d '{"request_id":"<X-Aiport-Request-Id of the answer>","vote":"up"}'
\`\`\`

\`vote\` is \`"up"\` or \`"down"\`. A \`score\` from 0 to 1 works instead, and 0.5 or
more counts as up. \`module\` names the module only when the call did not send
\`X-Aiport-Module\`, and it never replaces one the call sent. To change a vote,
send it again. The last one counts.

An app can rate only its own calls, and only calls still kept in request
history. Anything else gets \`404 request_not_found\`, and a malformed vote gets
\`400 invalid_feedback\`.

#### Read the report

The [Feedback](#feedback) page shows each Application's votes per module. Inside
each module it splits them by name sent, model and effort. With two or more
judged modules, it names the best and the weakest.

To improve the weakest module, change its Level, task name or effort under
Routing, or apply a Suggestion. Then watch its share of up-votes on the
Feedback page.

#### Suggestions from the votes

A Suggestion is a proposed change of model for one module, based on its votes.
aiport looks at a group of calls once it has at least 10 votes. A group is one
module's calls for one name, on one model and effort.

aiport reads the votes against your Levels and lists what it finds under
**Suggestions from the votes** on the Feedback page. There are three kinds:

| Kind | When | What aiport suggests |
| --- | --- | --- |
| **Better elsewhere** | Below 60% up, and a module of the same name (in another application, or this one) on another model gets at least 20 points more on the same protocol, also over at least 10 votes | That model, on its Connection. The best such evidence wins, and ties go to more votes. |
| **One Level up** | Below 60% up, the name sent is \`fast\` or \`balanced\`, and there is no such evidence | The next Level's model for that application (\`fast\` to \`balanced\`, \`balanced\` to \`best\`), untested for this module. |
| **Cheaper may do** | 90% up or more over at least 20 votes, and the name sent is \`best\` or \`balanced\` | The Level below's model for that application, untested for this module. |

Evidence comes only from the same module name, so aiport compares like with
like. A Level whose model already serves the module is not suggested, and calls
that named no module never get a Suggestion. There is at most one Suggestion
per application, module and name.

#### Apply a Suggestion as a module route

aiport never moves a module by itself. When you click **Apply**, aiport adds a
module route: that app's calls for that module that send that name go to the
suggested model. The route comes ahead of every other route, and the app's
other modules keep theirs.

Applying is one Change, with the reason from the votes included, and you can
revert it on Changes. Each app's panel on the Feedback page lists its module
routes with **Remove**. After applying, watch the module's share of up-votes.
The new model shows as its own row.

A module route has one model for every effort, with no per-effort models,
though a task name's fixed \`reasoning\` still applies. Its Connection must speak
the request's protocol, or the call is refused with \`400 incompatible_protocol\`.
In \`aiport.yaml\`, module routes are \`origins[].module_routes\`, each
\`{module, name, target, model}\`.

Scripts add one with \`PUT /api/v1/module-routes/{application}/{module}/{name}\`
and a body of \`{connection, model, why?}\`, and remove it with \`DELETE\` on the
same path. \`why\` goes into the Change as one line, cut at 300 bytes.

### Issues

An Issue is a problem aiport keeps seeing for one Application, such as calls to a model no provider serves. It sits at the top of Changes with the evidence, what aiport found out, and the ways to fix it: approve a fix in aiport, or hand a fix to the app's coding agent.

Handling one comes down to three steps:

1. Read the evidence and what aiport found out.
2. Choose a fix. **Approve** the fix in aiport, or click **Copy fix for the
   Coding agent** and paste it to the agent in the app's repository. You can
   also **Dismiss** the Issue.
3. Wait. aiport resolves the Issue by itself once the app's calls work again.

There is one Issue per Application and cause. When the problem repeats, aiport
updates that Issue's counts and time window instead of opening a second one.

#### The kinds of Issue

There are five kinds of Issue. aiport opens four of them by itself, and you
open the fifth from Activity.

An unknown model means the app asks for a model the provider does not have. A
failing Connection is a provider the app uses that keeps failing. An unmapped
Level is a Level the app asks for that is not set up for it. These three open
when a threshold under \`adapting:\` in \`aiport.yaml\` is reached.

A Catalog change opens when one of your Changes took away something the app
used. There is nothing to approve: it is a notice. A reported request is a
failed request you report yourself from Activity.

| Cause | Opens when | Fix in aiport | Fix in the Application |
| --- | --- | --- | --- |
| Unknown model (\`unknown_model\`) | A provider keeps answering \`404\` for a model the Application has no Alias for (\`adapting.unknown_model\`, default 5 times in 60 minutes) | An Alias for that Application to the Connection that serves the model | Send a model this gateway serves, or ask for the Alias |
| Failing Connection (\`failing_connection\`) | Its calls to one Connection keep failing (\`adapting.failing_connection\`, default 5 in 60 minutes) | Bind that Application to a working Connection | Back off and handle \`5xx\` and timeouts |
| Unmapped Level (\`level_unmapped\`) | Its calls for one Level keep being refused (\`adapting.level_unmapped\`, default 5 in 60 minutes) | Map the Level for that Application | Ask for the mapping, or send a mapped name |
| Catalog change (\`catalog_changed\`) | A Change takes away something it used | None needed | Send a name that still exists |
| Reported request (\`reported_request\`) | You report a failed request from Activity's **Details** | None | Whatever the failure and your note call for. The agent says it fixed it and the next call of that name and module confirms it |

The badge on Changes counts an Issue only while its fix waits for you.

#### What you can do with an Issue

On Changes, an Issue gives you a fix to approve, a fix to hand to the app's
coding agent, and a way to set it aside for a while.

Clicking **Approve** applies the fix in aiport as an ordinary Change you can
revert, and the Issue becomes **resolving**.

**Copy fix for the Coding agent** copies the fix in the Application, a ready
instruction for the coding agent in the app's repository. Paste it as it is. If
you want to look first, **Read the fix** shows the same text. The button is also
on Integrate, under the app's row. It is the same text aiport gives the app's
coding agent about its Issues, and it holds no prompts, keys or tokens.

Dismissing keeps the Issue quiet for 7 days, unless the evidence doubles. You
can dismiss any Issue that is not resolved or dismissed yet, wherever its fix
stands (waiting, approved, rejected or reverted). An approved fix stays applied
until you revert its Change.

Rejecting turns the aiport fix down and leaves the Issue as it is. The dashboard
offers only Approve and Dismiss, so Reject is for scripts: it is
\`POST /api/v1/issues/{id}\` with \`action: reject\`.

Issues are kept across restarts. They hold only Application ids, the cause,
model names, Connection ids, counts, request ids and fixed advice text. They
never hold prompts, responses, provider messages, keys or tokens.

#### An unknown model

The app asks for a model that the provider it reaches does not serve. For
example, docs-bot asks for \`llama3:8b\`, it is not bound to a Connection, and the
protocol default is a hosted Connection that does not serve that model.

The Issue shows the evidence: how many calls failed, the status and the
provider's error type (such as \`404 model_not_found\`), the time window, the
Connections the calls went to, and a few request ids to look up in Activity.

It also shows a diagnosis, which says which enabled Connections of that
protocol serve the model. The answer is exactly one, none, several, or unknown
when a listing failed. aiport finds out by asking each one for its model list.
That is the same free listing the connection check uses, kept for 2 minutes, so
nothing is generated and nothing costs money. A failed listing never counts as
"does not serve".

The fix in aiport is an Alias for that app only, with the same model name, to
the Connection that serves it. With exactly one serving Connection, the
Connection is filled in. With none, several or unknown, you choose the
Connection yourself (and the model, if it has another name there). **Approve**
applies it, and the Issue becomes **resolving**.

The instruction behind **Copy fix for the Coding agent** says what is wrong,
gives the evidence (counts, times, the Connections, example request ids) and
says what aiport found. It tells the agent what to change in the app: ask for
the Alias with a Change request, or send a model this gateway serves, listed for
that protocol. It also says how to confirm the fix. One call with the app's own
token answers \`200\`, and aiport then resolves the Issue by itself. Once the fix
in aiport is applied, the instruction says no code change is needed.

When the app's calls for that model succeed again, aiport marks the Issue
**resolved** and records how:

- \`aiport_fix\`: the Change you approved fixed it.
- \`application_fix\`: the app's coding agent fixed it.
- \`stopped\`: the calls succeeded, or stopped failing for a whole window,
  without a fix.

A fix is credited only when a call shows it works. If the app makes no calls for
that model for a whole window after you approved the fix, the Issue resolves as
\`stopped\`, not \`aiport_fix\`. Reverting the fix opens the Issue again and offers
the fix again. A resolved Issue also reopens when the error sticks again, and
only errors after it resolved count.

#### A failing Connection

The app's calls to one Connection keep failing. For example, docs-bot is bound
to a Connection that keeps answering \`502\`.

Gateway errors and timeouts count as failures. A provider answer below \`500\`
(an unknown model, a bad request), a cancellation or a request aiport refused
never counts. There is one Issue per Application and Connection, and only
failures to the Connection the app is bound to now count.

The evidence is how many calls failed, the last status, the time window and a
few request ids.

The fix in aiport binds that app alone to a working Connection. aiport proposes
the configured, enabled Connection with the most successful calls in the
window. You can choose another, but never the failing one. **Approve** applies
it as an ordinary Change you can revert, and other apps are not touched.

In the Application, the request itself is fine. The coding agent is told not to
retry in a tight loop, to back off and handle \`5xx\` answers and timeouts, and
that it can ask for the binding with a Change request of kind
\`bind_application\`.

It resolves itself when the app's calls succeed again, on that Connection or on
the one you bound it to (\`aiport_fix\` or \`stopped\`). It reopens if the failures
stick again. Reverting the binding opens it again with the fix offered again.

This fix is never applied automatically, even with automatic fixes on, because
moving a whole app to another Connection is always your decision. There is no
automatic fallback either. When a Connection is down or keeps answering with
errors, calls to it fail and the app gets the provider's own error. aiport never
sends calls where you did not choose.

#### An unmapped Level

The app keeps asking for a Level that is not mapped to a model for it. A Level
is a name like \`fast\`, \`balanced\` or \`best\` that stands for a model, and here,
for example, docs-bot sends \`best\` on the OpenAI protocol while no \`best\` is
mapped there, neither for docs-bot nor shared.

There is one Issue per Application, Level and protocol. If \`best\` is unmapped
on both protocols, two Issues open, each with its fix on its own protocol.

The evidence is how many calls were refused, the time window and a few request
ids, and the summary names the protocol.

The fix in aiport maps the Level for that app only, on the protocol its calls
were refused on. aiport proposes the Connection the app's other calls on that
protocol go to, which is its binding or else the protocol default. It leaves the
model to you, since it never makes up a model id. **Approve** applies it as an
ordinary Change you can revert, and as a Catalog change. A Connection of the
other protocol is refused. This fix is never applied automatically.

In the Application, the fix is to ask for the mapping with a Change request of
kind \`application_alias\` whose alias is the Level, or to send another Level or
task name the gateway maps.

It resolves itself when the app's calls for that Level succeed, and reopens if
the refusals stick again. Reverting the mapping opens it again with the fix
offered again. With the adapting loop off, nothing is counted and no Issue
opens.

#### A Catalog change

One of your Changes took away something the app was using, so the app should
switch to a name that still exists. For example, you remove a Connection and
leave docs-bot's Alias \`summarize\` without one. The Issue tells the app, and it
needs no decision from you.

aiport opens one Issue of cause \`catalog_changed\` per Change, for each app that
used what went away. "Used" means that, within request history
(\`telemetry.retention_days\`), the app sent the removed task name, or sent a
model name (not a task name) that went to the Connection its binding or the
moved protocol default pointed at.

The Issue says what changed (task names removed or moved, the binding or
protocol default and where it goes now), what the app can still use (its task
names and the Connections whose models it reaches), and how to confirm. Changes
and Integrate show it with nothing to approve, and it never lights the Changes
badge.

The app hears of it the way it hears of every Issue: through
\`GET /v1/aiport/issues\`, the MCP tool \`aiport_my_issues\`, the \`open_issues\`
count in the capabilities answer, and \`X-Aiport-Hint\` on a failing call that
sent a removed task name.

It resolves by itself when the app's calls succeed after the change (and after
its last failing call with a removed name). It also resolves at once when the
app's coding agent reports it fixed (\`aiport_issue_fixed\`), since there is
nothing in aiport to confirm, or when you revert the Change. You can dismiss it
too.

Not every app gets one. An app that never sent what went away, and the built-in
\`aiport\` Application, get the Catalog change recorded but no Issue. Apps the
Change took nothing from get nothing. A Change that is an Issue's own aiport fix
(or its revert) opens no Catalog Issue, because that Issue already says so. None
are opened while adapting is off, though the Catalog itself is still shown to
apps.

#### What the Application's coding agent sees

The app's coding agent hears about its Issues over the connection it already
has with aiport, using the app's own token. It sees what is wrong, the evidence
and the fix to make, and it reports back when it has fixed the code. Your
management token gets \`403\`.

The agent finds its Issues in four places:

- The MCP tool \`aiport_my_issues\`, or \`GET /v1/aiport/issues\` (and
  \`GET /v1/aiport/issues/{id}\` for one), lists the app's open and recent
  Issues. Each has what is wrong, the evidence, the fix in the Application
  (\`application_fix\`, the same text as **Copy fix for the Coding agent**), the
  root cause when you had the Issue explained (\`explanation\`), and whether the
  aiport fix is proposed, applied, rejected or reverted. An app never sees
  another app's Issues (their ids answer \`404\`), and never sees your notes.
- The capabilities answer (\`aiport_capabilities\`) shows \`open_issues\`, the
  number of its Issues that are open or resolving.
- A failing call that is one of an open Issue's failures carries an
  \`X-Aiport-Hint\` header naming that Issue, so the app's own logs point the
  agent to it. For an unknown model that is a provider \`404\` for the same
  model. For a failing Connection it is a \`5xx\`, timeout or connection error
  from that Connection, and for a Catalog change it is a failing call that sent
  a removed task name. Other failing calls, and aiport's own refusals that are
  not the Issue's cause, name no Issue. Issue hints are sent whether or not the
  app may request changes, because telling it about its Issues is not a Change
  request.
- The \`AGENTS.md\` line tells later sessions to check \`aiport_catalog\`,
  \`aiport_my_issues\` and \`aiport_guidance\` at the start.

When the agent has fixed the app's code, it says so with \`aiport_issue_fixed\`
(or \`POST /v1/aiport/issues/{id}/fixed\`). The Issue becomes **resolving** by
the Application fix, so you can see it is being handled. An aiport fix waiting
for you keeps waiting, and you can still approve it. Only \`open\` Issues can be
marked. Others get \`409 issue_not_open\`.

aiport does not take the agent's word for it. The Issue resolves as
\`application_fix\` only once the app's calls succeed again for the same cause.
For an unknown model, that means a successful call for that same model. For a
failing Connection, it means a successful call routed for that app after the
agent said so.

A successful call for another model does not count, and neither does silence.
If the error goes on, the Issue opens again and offers the aiport fix again.

### Explain

Explain tells you in a few sentences why an Application keeps failing, so you
can pick the right fix. Click **Explain** on an Issue, the problem aiport
noticed in an app's calls, and a model you choose writes the answer, or the
app's coding agent reads its code and writes it.

You find **Explain** on each Issue on the Changes page, and under the
Application's row on Integrate.

Who writes the explanation depends on the Application. Without a coding agent,
it is a Connection and model that you choose once in **Settings →
Explanations**. A Connection is a provider aiport can call, with its key: for
example, a local Connection serving \`llama3:8b\`. Until you choose one, Explain
is disabled and a hint says so. When the Application has a coding agent
connected, the agent reads the Application's code instead (see
[Explain by a Coding agent](#explain-by-a-coding-agent)).

In \`aiport.yaml\` the Connection and model are \`adapting.explain\`.

#### What is sent to the model

The model gets the Issue's evidence and nothing more. That evidence is:

- the cause, the status and the provider's error type
- the counts and the time window
- example request ids
- the model name
- the Connection ids
- aiport's diagnosis: which Connections list the model, and the models they
  list

Keys, tokens, the Application's id or name, prompts and provider messages are
never sent.

The only text in the evidence that an Application wrote is the model name it
asked for. That name is sent only when it looks like a model id: at most 128
characters, all letters, digits or \`. _ : / @ + -\`. Any other name is replaced
by "(a model name that is withheld because it contains unusual characters)".
Connection ids, model lists and the error type pass the same check. A coding
agent is told exactly the same.

#### While it writes, and after

The words appear as the model writes them. When it finishes, the explanation is
saved with the Issue, kept across restarts, and shows who wrote it.

**Explain again** writes a new one that replaces it. If writing fails or takes
longer than 2 minutes, nothing is saved and any earlier explanation stays. An
explanation that hits its length limit stops there. Only one explanation per
Application is written at a time.

#### Where the explanation goes next

The saved explanation becomes the "Root cause" paragraph at the start of "Copy
fix for the Coding agent". The Application's coding agent also gets it from
\`aiport_my_issues\` (and \`GET /v1/aiport/issues\`), both in that text and as the
\`explanation\` field.

#### In Activity, and what it costs

The call shows in Activity under the built-in Application \`aiport\`, with its
Connection, model, status and latency, like any other call. No configured
Application may use the id \`aiport\`. A remote Connection bills explanations to
its key, while a local one keeps them on this machine.

A script chooses who writes with \`PUT /api/v1/adapting/explainer\` and a body
such as \`{"connection": "local", "model": "llama3:8b"}\`. This is a revertable
Change, and sending both empty turns Explain off. \`POST
/api/v1/issues/{id}/explain\` streams JSON lines: \`start\` (who writes), \`text\`
(the next words), then \`done\` (the Issue with its saved explanation) or
\`error\`.

#### Explain by a Coding agent

When an Application has a coding agent connected, Explain lets that agent read
the app's code and write the root cause, usually naming the file and the change
to make. The agent works read-only in a throwaway copy of your repository, and
only when you click.

This needs three things: the Application has a coding agent, it has a
repository path, and the agent's ACP entry point (the program aiport talks to)
is installed. Then the **Explain** button's tooltip names the agent, and
clicking it starts an Agent run, which is simply a coding agent that aiport
starts on your click. Otherwise the Connection from Settings explains, as in
[Explain](#explain). To set an agent up, see
[Connecting a coding agent](#connecting-a-coding-agent).

##### What aiport cannot see

Read this before you use Agent runs.

aiport only sees the permission requests an agent sends it. An agent's own
settings can let it read, edit or run commands without asking. Those actions
never reach aiport: it cannot allow them, deny them or log them.

aiport limits what such settings can do, in layers:

- Every run works in a worktree made for it, never in your working copy.
  Explain's worktree is thrown away when the run ends, and Fix it's is shown to
  you as a diff. An edit nobody was asked about lands only there.
- aiport picks the most restrictive session mode the agent offers (see
  [Session modes](#session-modes)), and does not run a session that can only
  act without asking.
- Fix it never allows edits to agent or git configuration in the worktree, and
  any edit the agent reports ends "Allow until the next edit".
- Before a run, the run panel warns you when the repository's own agent
  settings allow actions without asking. It looks for:
  - \`.claude/settings.json\` or \`.claude/settings.local.json\` with
    \`permissions.allow\` entries, or a \`defaultMode\` such as
    \`bypassPermissions\` or \`acceptEdits\`
  - \`.gemini/settings.json\` with \`autoAccept\`, or an approval mode such as
    \`yolo\`
  - \`.codex/config.toml\` with \`approval_policy = "never"\` or
    \`sandbox_mode = "danger-full-access"\`

  aiport only reads these files. It never changes them.

What aiport cannot check: your user-level agent settings (for example
\`~/.claude/settings.json\`, \`~/.gemini/settings.json\` or \`~/.codex/config.toml\`)
and any managed settings. A command those settings allow runs without aiport
asking you. It runs with your own rights, in the run's worktree, or elsewhere
if the agent chooses.

**Warning:** do not auto-allow commands or edits in your user-level agent
settings for repositories you use with aiport. Check the warnings in the run
panel.

##### How a run starts

A run starts only when you click. It never starts by itself, not even with
automatic fixes on, and never with adapting off.

aiport starts the ACP entry point it found by its absolute path, never through
a shell. The agent gets a minimal environment: \`PATH\` (the agent's own
directory first, then the absolute directories of aiport's \`PATH\`, so an
npm-installed adapter finds \`node\`), \`HOME\`, \`TMPDIR\`, \`LANG\`, \`LC_ALL\`, \`USER\`,
\`LOGNAME\` and, for Claude Code, the variables in
[Where the run's model calls go](#where-the-runs-model-calls-go). It never gets
aiport's tokens or provider keys, and it has no MCP connection to aiport.

##### A throwaway copy, never your working copy

aiport makes a git worktree of the repository for the run, in
\`aiport.yaml.worktrees/\`. It is detached at your current commit (\`HEAD\`, no
branch), and aiport removes it when the run ends. The agent reads that copy and
never your working copy, so even an edit that the agent's own settings allow
without asking cannot touch your files.

Your uncommitted changes are not in that copy. The repository needs at least
one commit.

##### What the agent is told

The agent gets the Issue's evidence (the same as in [Explain](#explain)) and is
told that the run is read-only.

##### Read-only: what the agent may do

aiport answers every permission request itself, without asking you. Reading and
searching inside the run's copy is allowed. Everything else is denied:

- every edit, delete or command
- anything outside the run's copy, including your own working copy
- a permission request that names a session aiport did not open

aiport only ever answers "once", so nothing is remembered in the agent's own
settings.

Every path a request names counts, not only its locations. That includes the
paths anywhere in its raw input: \`path\`, \`file_path\`, \`cwd\`, a move's
\`destination\`, the files of a patch, and the base of a search pattern such as
\`/etc/**\` or \`../*.py\`. One path outside the copy (or in your home directory,
\`~\`) denies the whole request. A read or search that names no path at all is
denied too. The same rules apply to Explain and Fix it.

##### Session modes

Some agents offer modes. Claude Code's adapter, for example, offers \`default\`,
\`acceptEdits\`, \`plan\` and \`bypassPermissions\`. aiport switches the session to
the most restrictive suitable mode.

For Explain that is plan or read-only. For Fix it, it is a mode that asks
before acting, such as \`default\`.

It never picks a mode that bypasses permissions or auto-accepts. If the session
starts in such a mode (from your own settings) and the agent offers no other,
aiport does not run it. The run panel names the mode chosen.

##### The run panel

The run panel under the Issue shows the agent's name, where its model calls
go, its steps (thoughts, files read, permissions allowed or denied) and its
words as they come.

**Stop** ends the run at any time. The agent is asked to cancel, and then its
whole process group is ended. Nothing is saved from a stopped run.

Closing the tab or leaving the page does not stop the run. It goes on, saves
its explanation when it ends, and the run panel picks it up again when you
come back.

##### Limits

Only one Agent run per Application goes at a time. An Explain run stops after
10 minutes (\`agent_runs.explain_minutes\`), and nothing is saved.

##### What is kept

The explanation is saved with the Issue and replaces any earlier one, including
one a Connection wrote. It is passed to the Application's coding agent as in
[Explain](#explain).

The run's log (its steps and words) is kept in memory only, while the Issue is
open or resolving. It is never written to disk, and it is gone when the Issue
closes or aiport restarts.

##### Where the run's model calls go

The agent itself calls a model while it works. You choose where those calls go
in **Settings → Default coding agent → Model calls of Agent runs**: an
Anthropic-protocol Connection and, optionally, a model.

Until you choose one, every run uses the agent's own login, and the run panel
says so. aiport never sends a run's calls to the Anthropic protocol default
unless you chose it for Agent runs.

Each agent handles this differently:

- Claude Code is pointed at aiport through its environment only:
  - \`ANTHROPIC_BASE_URL\`: aiport's address on this machine (the listen port on
    \`127.0.0.1\`, or the listen address when that is a loopback address)
  - \`ANTHROPIC_AUTH_TOKEN\`: a credential made for this run
  - \`ANTHROPIC_MODEL\`, when you chose a model

  That credential only authenticates model calls, as the built-in Application
  \`aiport\`. It cannot reach the dashboard, the MCP endpoint or any
  Application's data. It is never written down or logged, and it stops working
  when the run ends. The calls show in Activity under \`aiport\` with their
  Connection. With no Connection chosen for Agent runs, or when the chosen one
  is not reachable, Claude Code uses its own login and the run panel says so.
- Codex uses its own login. Pointing it at aiport would mean editing its
  configuration file, which aiport never does. The run panel says so.
- Gemini CLI uses its own login too, because aiport does not speak the Gemini
  API yet. Again, the run panel says so.

\`\`\`yaml
agent_runs:
  default_agent: claude
  connection: claude-hosted   # an Anthropic-protocol Connection
  model: claude-model-id      # optional
  explain_minutes: 10
\`\`\`

Over the API, \`PUT /api/v1/agents/calls\` with \`{"connection", "model"}\` sets
where model calls go (a revertable Change). \`GET /api/v1/runs\` lists the runs
going, and \`POST /api/v1/runs/{id}/stop\` stops one. \`GET
/api/v1/issues/{id}/run\` gives the Issue's last run and its log, while the
Issue is open. For an Agent run, the explain stream adds \`run\` events (the run:
its id, agent and model calls) and \`step\` events, and its \`error\` codes include
\`run_stopped\` and \`explanation_timed_out\`.

#### Fix it

Fix it lets the Application's coding agent change the app's code on a branch of
its own. It ends with a diff that you review, then keep or discard. Nothing is
committed until you click **Keep the branch**.

Fix it shows on an open Issue when the Application has a coding agent and a
repository path, as for [Explain by a Coding agent](#explain-by-a-coding-agent).
Otherwise a one-line hint under the Issue says what is missing, for example
"Add the repository path of docs-bot".

It is offered only when the cause can be in the app's code: the app asks for a
model name that no Connection serves. When a Connection is failing, the Issue
says instead that the cause is not in the app's code, and points to the aiport
fix (binding the Application to a working Connection).

##### Its own worktree and branch

aiport makes a git worktree of the repository on a new branch named for the
Issue, \`aiport/fix-<issue id>\`, with \`-2\`, \`-3\` and so on if that branch
exists. It starts at your current commit (\`HEAD\`).

The worktree lives next to the configuration, in \`aiport.yaml.worktrees/\`, not
inside your repository. Your working copy, its uncommitted changes and your
branch are never touched, and your uncommitted changes are not in the worktree
either. The repository needs at least one commit.

##### What the agent may do

aiport answers every permission request itself. Inside the worktree the agent
may read, search, and edit, create, delete or move files. Symbolic links are
followed, so a link that points out of the worktree counts as outside.

Some files are never allowed, at any level of the worktree and in any letter
case:

- the agents' own configuration: \`.claude/\`, \`.codex/\`, \`.gemini/\`,
  \`.mcp.json\`, \`CLAUDE.md\`, \`CLAUDE.local.md\`, \`AGENTS.md\`, \`GEMINI.md\`
- git's configuration: \`.git/\`, git hooks in \`.husky/\` or \`.githooks/\`,
  \`.gitattributes\`, \`.gitmodules\`

So a branch you keep cannot change how your agents or git behave.

A command inside the worktree waits for your answer (see
[Approving commands](#approving-commands)). Anything outside the worktree is
denied without asking, including your own copy of the repository and a command
that says it runs elsewhere.

aiport only ever answers "once" to the agent, so nothing is remembered in the
agent's own settings. The agent starts as it does for Explain: by absolute
path, with the minimal environment, only on your click.

##### What the agent is told

The agent gets the Issue's evidence, the saved explanation if there is one, and
the branch it works on. It is told that it must not commit, merge or push.

##### While it runs

The run panel shows its steps and the permissions allowed or denied, with
**Stop**. The run goes on if you leave the page.

A run stops after 30 minutes (\`agent_runs.fix_minutes\`). Whatever it changed by
then stays in its worktree and waits for your review with **Keep the branch**
and **Discard**, as after a finished run (state \`timed_out\`).

Some runs leave nothing behind. Stop, a run that fails, and a run that changed
nothing (including one that timed out or went unanswered before changing
anything) all remove the worktree and the branch. The Issue stays open and
shows why the run ended.

If aiport stopped or crashed while a Fix it run was going, the run is marked
stopped the next time aiport starts, and its worktree and branch are removed.
The same happens to the throwaway worktrees of Explain runs that were going.
Only directories inside \`aiport.yaml.worktrees/\` and branches named
\`aiport/fix-…\` are ever removed.

##### What is kept

The branch name, the commit it started from, the diff summary and the changed
files are saved with the Issue, across restarts. Once you keep the branch, so
is the commit aiport made. The diff itself is read from the worktree or the
branch when you look at it. The run's log is kept as for Explain, in memory
only.

##### Approving commands

When the agent asks to run a command (tests, a build), the Issue shows it and
waits for your answer.

What you see is exactly what runs:

- The whole command, exactly as the agent sent it, as plain text. It is the
  agent's command, not aiport's.
- It is never cut. A long command scrolls.
- Characters that would not show as themselves are written as escapes:
  control characters, invisible and direction-changing characters (such as
  U+202E, the right-to-left override), and unusual spaces.
- A list of arguments is shown shell-quoted (\`echo 'a b'\`).
- A command longer than 16 KiB is denied without asking, never shown in part.

You have three answers:

| Answer | What it allows |
| --- | --- |
| **Allow once** | This one time. |
| **Allow until the next edit** | This time, and every time the run asks for the exact same command again, until the agent changes a file (whether it asked to or not). "The same" means the same list of arguments, or the same string, in the same directory: \`echo 'a b'\` is not \`echo a b\`. Another command, or the same one in a later run, asks again. |
| **Deny** | Nothing. The agent is told no and goes on without it. |

An edit can change what a command does, which is why "Allow until the next
edit" asks again after one. But aiport only sees the edits the agent asks about
or reports. A command you allowed that changes files itself (a code generator,
a formatter) does not end the allowance, so choose **Allow once** for commands
that write files.

The agent runs an allowed command itself, in its own process, in the
worktree's directory. aiport only answers the permission. Whatever the command
changes in the worktree shows in the diff.

While a command waits, the Changes badge in the Monitor menu counts it. A
command left unanswered for 5 minutes stops the run. What it changed so far
stays in its worktree, and the Issue shows it for review with **Keep the
branch** and **Discard**, as after a finished run (state \`stopped_waiting\`).

##### Reviewing the diff

When the agent ends its turn, the Issue shows **Review the change**: the
branch, the summary (for example "2 files changed, 3 insertions(+), 1
deletion(-)") and the whole diff. Nothing is committed yet.

**Keep the branch** makes aiport commit the diff on the new branch as one
commit, authored by \`aiport <aiport@localhost>\`, and remove the worktree. This
is the only commit aiport ever makes, and only when you click Keep. It goes on
the new branch, never on yours, with hooks off. The branch stays in your
repository for you to look at, merge or delete, since aiport never commits to
your branch, merges or pushes. The Issue becomes resolving, as an Application
fix, and resolves when the Application's calls succeed again, that is, after
you have merged and deployed the change. From then on, "Copy fix for the
Coding agent" names the branch.

**Discard** removes the worktree and deletes the branch. The Issue stays open,
and you can Fix it again.

##### git, safely

aiport finds git once, at startup, by absolute path: \`/usr/bin/git\` when it is
there, otherwise the first \`git\` in its \`PATH\`. The startup output names it.

It runs git with a minimal environment, without your global or the system's git
configuration and attributes (\`GIT_CONFIG_GLOBAL=/dev/null\`,
\`GIT_CONFIG_NOSYSTEM=1\`, \`GIT_ATTR_NOSYSTEM=1\`,
\`core.attributesFile=/dev/null\`). Hooks, fsmonitor, commit signing and
external diff programs are turned off.

What the repository itself sets in its own \`.git/config\` and \`.gitattributes\`
still applies, for example a clean or smudge filter it defines. It is your own
repository, so check those files before you use Fix it on a repository you did
not set up.

aiport removes a worktree only when its directory is directly inside
\`aiport.yaml.worktrees/\`. Any other recorded path it finds is left alone and
logged.

\`\`\`yaml
agent_runs:
  fix_minutes: 30
\`\`\`

Scripts drive Fix it through these routes:

- \`POST /api/v1/issues/{id}/fix\` starts a run and answers \`202\` with the run.
  It answers \`409\` with \`fix_unavailable\` and the hint, \`fix_not_for_cause\`
  for any cause but an unknown model, \`fix_waiting\` while an earlier diff
  waits for you, \`agent_run_running\`, \`issue_not_open\` or \`worktree_failed\`.
- \`GET /api/v1/issues/{id}/fix\` gives \`branch_fix\` and the \`diff\`. \`POST
  /api/v1/issues/{id}/fix/keep\` and \`…/fix/discard\` keep or discard it.
- \`GET /api/v1/issues\` says per Application, in \`fix_by\`, whether Fix it is
  available: with which agent and time limit, and for which \`causes\` (with
  \`cause_hint\` for the others), or the hint. Each Issue carries its
  \`branch_fix\`.
- \`GET /api/v1/approvals\` lists the commands waiting for you, each with \`id\`,
  \`run\`, \`application\`, \`issue\`, \`command\` and \`expires_at_ms\`.
- \`POST /api/v1/approvals/{id}\` answers one, with \`{"answer": "allow_once"}\`,
  \`"allow_run"\` (allow until the next edit) or \`"deny"\`. Once it is answered
  or its run has ended, this gives \`404 approval_not_found\`.
- \`GET /api/v1/runs\` gives each run's \`waiting\` count, and \`GET
  /api/v1/issues/{id}/run\` the Issue's waiting \`approvals\`.

These are management routes, never open to an Application token, and answering
needs the dashboard's CSRF header.

### Connecting a coding agent

Connect a coding agent to an Application so aiport can look at the app's code,
on your click. Explain reads the code and tells you the root cause, and Fix it
makes the change on a new branch and shows you the diff to keep or discard.
Explain only reads, nothing runs without your click, and aiport never commits
to your branch, merges or pushes.

Without an agent, an Issue still offers the fixes aiport can see from its side.
Explain still works too, more briefly: it uses only the Issue's evidence, with
the Connection and model chosen under **Explanations** in Settings.

#### Set it up

aiport can start a coding agent installed on this machine to work on an
Application's Issues, only when you click. These are called Agent runs. Each
Application can have its own agent (**Claude Code**, **Codex** or **Gemini
CLI**) and the path of its git repository.

1. Open step 6 of the onboarding wizard, or the **Agent** column on Integrate.
2. Choose the agent. If it is missing, install it first (see below).
3. Enter the repository path.
4. Click **Save**.

Each save is one revertable Change on the Changes page.

**Default coding agent** on the Settings page is the agent for every
Application that has none of its own.

#### What aiport looks for

aiport talks to agents through the Agent Client Protocol (ACP).

| Agent | What aiport needs |
| --- | --- |
| Gemini CLI | Nothing extra: it speaks ACP itself (\`gemini --acp\`). |
| Claude Code | The adapter \`claude-agent-acp\` (npm package \`@agentclientprotocol/claude-agent-acp\`). |
| Codex | The adapter \`codex-acp\` (npm package \`@agentclientprotocol/codex-acp\`). |

aiport finds programs by absolute path. It looks in the absolute directories of
its own \`PATH\` and in the usual install places (\`~/.local/bin\`,
\`~/.npm-global/bin\`, \`/opt/homebrew/bin\`, \`/usr/local/bin\` and a few more). It
does not run them to look.

#### Nothing is installed for you

When an agent or its adapter is missing, the step shows the exact command to
install it, such as \`npm install -g @agentclientprotocol/codex-acp\`. Run the
command yourself, then refresh the page.

An agent that is not installed cannot be chosen.

#### The repository path

The path must be absolute and must be the top directory of a git repository
(where \`.git\` is). Otherwise it is refused, with the reason. aiport stores it
with symbolic links resolved, and Agent runs never work outside it.

#### In the configuration file

In \`aiport.yaml\` these are \`agent\` and \`repository\` on the Application under
\`origins\`, and \`agent_runs.default_agent\`:

\`\`\`yaml
agent_runs:
  default_agent: claude
origins:
  - id: docs-bot
    name: Docs Bot
    token_env: AIPORT_APP_DOCS_BOT_TOKEN
    agent: gemini
    repository: /home/you/src/docs-bot
\`\`\`

To see what aiport found from a script, call \`GET /api/v1/agents\`. It lists
\`installed\`, \`ready\`, the ACP entry point, and \`install\` when something is
missing.

### What waits for your decision

Suggestions and Change requests are changes to aiport's configuration that
someone else proposes. Each one waits on the Changes page until you approve,
dismiss or reject it, and nothing in this list changes aiport until you decide.

There are two kinds.

Suggestions come from aiport itself, from patterns in request history. Each
shows its evidence: how many requests, over which time window.

- aiport suggests pairing when requests keep arriving with a missing or unknown
  token (\`adapting.unknown_token\`). This one is informational: open the
  Integrate page to pair the application, or click **Dismiss**.
- Two other fixes come as [Issues](#issues) instead. One is an Alias (another
  name for a model) for a model one Application keeps getting \`404\` for. The
  other binds an Application to a working Connection when its calls to one
  Connection keep failing.
- The thresholds are under \`adapting:\` in the configuration.
- Suggestions from the votes are different. They live on the Feedback page
  (see [Suggestions from the votes](#suggestions-from-the-votes)).

Change requests come from Applications that you allowed to request changes.

- An Application names a task, not a provider, for example the alias
  \`review-code\`. You decide which Connection and model serve it.
- It may add advice, such as "suggests big-model on hosted", with its reasons.
  The advice pre-fills the form. You can choose otherwise now, or change it
  later without touching the Application's code: in \`origins[].aliases\` in the
  configuration file, or with a new Change request.

#### Reading a Change request

Each Change request shows what it asks for. **Details and reasons** opens the
rest: the full details, why it is needed, the argument for the Scope asked for,
and who filed it (as the request states).

When several Applications send identical requests, they are merged into one
item, each with its own argument. One decision answers them all.

Request text is plain text from the Application. It is shown as is and never
sent to a model. Each field is at most 2,000 characters, and 6,000 in total.

#### Waiting, Accepted, Rejected

The **Waiting / Accepted / Rejected** filter on **Suggestions and Change
requests** picks what the list shows.

**Waiting** is what needs you. **Accepted** and **Rejected** show past
decisions, most recent first, with who decided, when, and the note. Rejected
includes dismissed Suggestions.

Decided items are kept as long as request history
(\`telemetry.retention_days\`). An accepted item's Change stays in **Applied
changes**, where you can revert it.

Kinds of request aiport cannot apply are kept as "not supported" and counted
under **Asked for, not supported yet** (kinds and counts only). They need no
decision.

#### Scope

Every decision has a Scope: it applies **only** to the requesting Application,
or to **every Application**. The requester argues for one, and you choose. For
an Alias, every Application means a shared Alias. For a binding, it means the
protocol's default Connection.

#### Approve, Dismiss, Reject

**Approve** applies the Change at once. For an Alias you confirm the alias
name, the Connection and the model, and for a binding, the Connection. You can
add a note of up to 2,000 characters.

When you **Dismiss** a Suggestion, aiport stays quiet about it for 7 days,
unless the evidence doubles.

When you **Reject** a Change request, your note goes back to the Application,
which sees it when it lists its requests.

#### Allowing Change requests

An Application may request changes only after you allow it. You can do that in
any of three places:

- tick **May request changes** when you approve its pairing
- use the toggle on Integrate
- use the toggle in the Applications table on Routing

Until then it gets \`403 change_requests_not_allowed\`.

Allowed Applications also get hints when an error is one a Change could fix:
\`error.hint\` on \`target_unconfigured\` and \`target_disabled\`, and the
\`X-Aiport-Hint\` header on a provider \`404\` for the requested model.

Change requests are limited to 20 per Application per hour (\`429
change_requests_rate_limited\`). A request that a policy refuses does not count
toward that limit.

The coding-agent brief teaches an Application to write a good Change request
with these fields:

- \`kind\`, from \`change_kinds\`: \`application_alias\` or \`bind_application\`
- \`need\`: one line
- \`details\`
- \`why\`: the error code, the \`X-Aiport-Request-Id\`, what it tried
- \`scope\`: \`application\` or \`every_application\`, and \`scope_case\`
- optional: \`filed_by\`, \`alias\`, \`prefer_connection\` and \`prefer_model\`

A \`400 invalid_change_request\` lists every problem in \`error.problems\`. The
Application reads the answer with \`aiport_my_requests\`: \`pending\`, \`accepted\`,
\`rejected\` (with your note) or \`not_supported\`.

### Applied changes and Revert

Applied changes lists every configuration Change made through aiport, newest
first, with who made it. **Revert** undoes any one of them, and the list
survives restarts, so you can always get back to an earlier configuration.

A Change is made by you, by an approved Suggestion or Change request, or by
**aiport (automatic)** for an automatic fix.

**Revert** restores the configuration file exactly as it was before that
Change, and makes it live at once. If the file was changed after that Change,
revert the newer Changes first: aiport refuses rather than overwrite them.

Applied Changes are never pruned.

#### Automatic fixes

Automatic fixes let aiport apply an Issue's aiport fix itself, without waiting
for you, in a few safe cases. They are off by default.

To turn them on, use the **Automatic fixes** switch under Applied changes, or
set \`adapting.auto_fix: true\` in \`aiport.yaml\`. The switch itself is a Change
you can revert.

aiport applies a fix automatically only when all of these hold:

- The diagnosis is certain: exactly one enabled Connection serves the model.
  Never when none, several or an unknown number do.
- The error is sticky, meaning the Issue's threshold (\`adapting.unknown_model\`)
  was reached.
- The fix touches only that Application, as an Alias for it alone. Binding a
  whole Application to another Connection, Level mappings, shared Aliases,
  protocol defaults and anything for every Application are never automatic.
- You have not refused that fix (see below), and the adapting loop is on.

For example, docs-bot keeps getting \`404\` for \`llama3:8b\`, and only the local
Connection serves it. aiport adds the Alias \`llama3:8b\` → local for docs-bot,
and the Issue becomes resolving.

The Change is recorded as made by **aiport (automatic)**, and Applied changes
shows it as "aiport fixed: …" with **Revert**. The Issue shows "aiport fixed
this automatically", with its own Revert. Guardrail Extensions judge automatic
fixes just as they judge your approvals.

If you revert an automatic fix, aiport records that you said no. It never
applies the same fix automatically again for that Application and cause, also
after a restart. The Issue opens again and its fix waits for your decision, and
you can still approve it.

A fix you rejected is never applied automatically either, even if you turn
automatic fixes on later. You can still approve it yourself.

To remember these decisions, an Issue that holds a fix you reverted or rejected
is kept beyond request-history retention (\`telemetry.retention_days\`), even
once it is resolved or dismissed. It holds no more than any other Issue: ids,
model names, counts.

#### The sound

The speaker button next to **Applied changes** plays a soft sound when
something new waits in Changes. It is off by default, and the setting is kept
in this browser only.

### Compress requests

An application can ask aiport to shrink a request before it goes to the
provider, and aiport measures whether that makes the answers worse.
Compression is off unless the application asks for it, request by request,
with the \`X-Aiport-Compress\` header.

To turn it on for a request, the application sends the methods it wants, for
example:

\`\`\`
X-Aiport-Compress: whitespace, json
\`\`\`

| Method | What it does |
| --- | --- |
| \`whitespace\` | Removes trailing spaces, normalises line endings, and turns two or more blank lines in a row into one. Indentation is kept. |
| \`json\` | Rewrites text that is entirely JSON (typical tool results) in compact form. |
| \`dedupe\` | When a block of 1 KiB or more repeats, earlier copies become a short note and the latest copy stays. Lossy. |

\`all\` picks every method. \`off\`, or no header, picks none.

Only message text is changed. Tool definitions, schemas and metadata are never
touched, and if compression fails or saves nothing, aiport sends the original
bytes.

To see whether it is worth it, open the [Compression](#compression) page. It
shows the savings and a verdict for each combination of methods. The full
description, including the evaluator, is in [compression](compression.md).

An unknown method returns \`400 unknown_compression\`. Every compressed response
carries \`X-Aiport-Compression\` with the methods, the status (\`status=applied\`,
\`no_gain\` or \`failed\`), and the bytes before and after.

### Add Extensions

An Extension is a program of your own that helps decide on Change requests. Use
one to enforce your own rules without changing aiport.

There are two kinds:

- A **policy** runs when a Change request arrives and may refuse it. The
  Application then gets \`403 change_request_refused\` with the reason.
- A **guardrail** runs when you approve, before the Change is applied, and may
  block it (\`guardrail_blocked\`). Changes then shows which Extension blocked
  it and why, and the item keeps waiting.

To add one:

1. List it under \`extensions:\` in \`aiport.yaml\`, with an absolute path.
2. Restart aiport.
3. Check that it appears in the **Extensions** panel on Changes.

aiport runs only programs you listed, each with a timeout. Anything other
than an explicit allow counts as a deny. The input and output protocol is in
[Extensions](extensions.md).

### Suggest to aiport and Share upstream

You can send ideas back to the aiport project in two ways. Both give you a
draft to check first, and nothing is ever sent on its own.

**Suggest to aiport** sits under **Asked for, not supported yet**. Use it for a
kind of Change that Applications asked for and aiport cannot apply.

**Share upstream** sits in the **Extensions** panel of Changes. Use it to offer
an Extension you listed. The draft names the hook and the input protocol, never
the Extension's name, path or settings. You add the program yourself, after
removing anything private.

A draft contains kinds, hooks and anonymous counts only. It never holds
configuration, keys, paths, Application names or request text. One thing to
check: a kind named by an Application is in that Application's own words.

You can edit both the title and the body, and every way of sending uses your
edited text:

- **Open on GitHub**, when \`upstream.repository\` is set.
- **File with gh…**: after you confirm, aiport runs \`gh issue create\` with
  your own gh login.
- **Copy text**.

### Turn the adapting loop off

The adapting loop is the part of aiport that watches request history and
proposes fixes: Suggestions, Issues, Change requests and Explain. You can turn
it off and keep everything else working.

1. Add this to \`aiport.yaml\`:

   \`\`\`yaml
   adapting:
     disabled: true
   \`\`\`

2. Restart aiport.

With the loop off, aiport finds no Suggestions and opens no Issues, so Changes
shows nothing new. Change requests and \`/v1/aiport/issues\` answer \`409
adapting_off\`, and the Issue tools are not offered over MCP. Capabilities carry
no \`open_issues\`, no hints are sent, and Explain and Agent runs are
unavailable. \`/_connect\` and both briefs mark change requests, Issues, their
MCP tools and \`X-Aiport-Hint\` as off, and say why.

Everything else keeps working:

- Calls are sent to providers exactly as before.
- Revert works on earlier Changes.
- The Catalog (the list of names each Application can send) is still shown
  to Applications: \`/v1/models\`, \`X-Aiport-Catalog\`, \`/v1/aiport/catalog\` and
  \`aiport_catalog\`. Catalog changes are still recorded, but no Catalog-change
  Issues are opened.
- Levels, guidance, the communication standard and votes.

The detector thresholds and the check interval (\`adapting.interval_seconds\`)
are in [configuration](configuration.md).

### Use a name instead of 127.0.0.1

You can open the dashboard at \`http://aiport.localhost:8787/\` instead of the
IP address, because browsers send any \`*.localhost\` name to this machine by
themselves. The name is for your browser only. Applications and coding agents
keep using the numeric address.

1. Add this to \`aiport.yaml\`:

   \`\`\`yaml
   hostnames: [aiport.localhost]
   \`\`\`

2. Restart aiport.

The dashboard then works at the name straight away, and every page has its
address there, for example \`http://aiport.localhost:8787/integrate\`.

A name must be lowercase, with no port, and must be either \`localhost\` or end
in \`.localhost\`.

aiport still listens on \`127.0.0.1\`. It still refuses every name you did not
list, with \`403 invalid_host\` or \`invalid_origin\`.

Applications and coding agents are always given \`http://127.0.0.1:8787\`, even
when hostnames are set. That covers the agent instruction, both briefs, the
communication standard, \`/_connect\`, the \`.aiport\` lines on a token card and
the MCP entry. The number stays because a name causes three problems that it
does not:

- a name needs a lookup the system resolver may not do
- it resolves to \`::1\` first, and aiport listens on IPv4
- a coding agent's sandbox sends it through its network proxy, which refuses
  it

## A tour of the dashboard

The dashboard is where you run aiport day to day. It shows your traffic, holds
your Connections and routing, connects your Applications, and lists what
waits for your decision. The Overview page shows the traffic at a glance and
one line for each thing that needs you.

The sidebar sorts the pages into three groups. **Monitor** holds the ones you
use every day: Overview, Activity and Changes. **Configure** holds Integrate,
Providers, Routing, Agents, Compression and Feedback, and **Workspace**, at
the bottom, holds Tasks and Settings.

The bottom of the sidebar also shows the address aiport listens on, and
**Sign out** (or "Authentication off" when sign-in is turned off). In the
toolbar at the top right are the help button, which opens this manual, the
**Light** / **Dark** / **System** appearance switch, and Refresh (on Tasks,
also the chime).

The dashboard refreshes itself every 10 seconds while it is open. Refresh
fetches everything at once, which helps right after you edited the
configuration file by hand.

The ⓘ next to a page title or panel opens the start of the matching manual
section, with a link to the whole section.

Each page has its own address (\`/integrate\`, \`/changes\`, \`/manual\` and so on).
Reloading stays on the page, Back and Forward move between pages, and a link
opens the page it names.

| Page | What it is for |
| --- | --- |
| [Overview](#overview) | Traffic at a glance, and one line for each thing that needs you. |
| [Activity](#activity-and-history) | Every recorded request, with why it failed. |
| [Changes](#changes) | Issues, decisions, and every Change with Revert. |
| [Integrate](#the-integrate-page-and-the-two-briefs) | Your Applications, onboarding and the instructions for coding agents. |
| [Providers](#connections-and-keys) | Your Connections and their keys. |
| [Routing](#routing-and-protocol-defaults) | Protocol defaults, Levels, the Routing map and the Applications table. |
| [Agents](#agents) | Coding agents launched through \`aiport agent\`. |
| [Compression](#compression) | Opt-in request compression and whether it hurts answers. |
| [Feedback](#feedback) | Votes per Application, module, model and effort, and Suggestions from them. |
| [Tasks](#tasks) | A queue of work for coding agents. |
| [Settings](#settings) | Sign-in, the default coding agent, Explanations and the configuration file. |

### Overview

Overview shows the shape of your AI traffic at a glance, and tells you when
something needs you.

The traffic part shows request counts, average and P95 latency, error rate,
hourly traffic, the provider distribution (with each provider's logo) and the
latest requests.

Below that, you get one line for each thing that needs you:

- Something waits for your decision. Use **Review in Changes**.
- An Application's coding agent has not confirmed the current guidance
  version. Use **See Guidance in Integrate**.
- The latest request on a route failed because of something in your
  configuration. The line gives the reason and the fix, and goes away once a
  request there succeeds.

A banner says when request history is off, kept for this session only, or
incomplete.

### Activity and history

Activity lists every recorded request, so you can see who called what, with
which name and effort, what served it and why it failed. aiport records only
metadata: time, route, Application, model, effort, module, status and
latency. It never records prompts, responses or headers.

You can filter by period, Application (**Origin**), Connection (**Target**),
outcome, and the name sent (**Name**). **Load older requests** goes further
back.

**Details** opens one request right under its row. It shows, in order, why it
failed and what fixes it, then what the Application asked for (name, module,
effort), what served it (Connection, model, provider, route), the result
(outcome, latency, the user's vote), and finally its id and time.

#### Hand a failure to the Application's coding agent

Under a failed request from an Application, **Report to the Application's
coding agent** opens an Issue (a problem aiport tracks for one Application)
with the cause \`reported_request\`. The Issue holds the request's metadata,
aiport's explanation and your note.

1. Open **Details** on the failed request.
2. Paste the error the application showed into the note. Request history
   keeps no provider messages, so this is the only place the error text comes
   from.
3. Select **Report to the Application's coding agent**.

The coding agent reads the Issue at the start of its next session
(\`aiport_my_issues\`). Until then, its next failing call of the same kind
(same status, error, name and module) names the Issue in \`X-Aiport-Hint\`, and
reporting another failure of that kind adds it to the same Issue rather than
opening a new one.

The Issue resolves after the agent says it fixed it and the next call with
that name and module succeeds. Another failure of that kind reopens it.

aiport cannot talk to a coding agent session that is already running. To have
aiport run a coding agent on the Issue now, open it in Changes and use **Fix
it**.

#### The live gauge

The gauge at the top of every page shows, live, whether aiport answers and
which Applications are calling. It reads request metadata only.

**Listening** is green while the gateway answers the dashboard. It reads
**No reply** when it does not, and **Signed out** when your session has
ended.

Next to it is one light per Application. A light is green when that
Application called in the last two minutes, red when its call in the last two
minutes failed, and faint when it called in the last hour. A level meter
lights up with each call as it passes, red at the top for a failure.

#### Turning request history on

Set \`telemetry.enabled: true\`. aiport then records request metadata in a
private SQLite file next to the configuration (\`aiport.yaml.db\`, mode 0600).

Recording never slows down calls to providers. If records are dropped or fail
to save, the banner on Overview and Activity says how many. And if the
database cannot be opened at all, aiport says so and keeps the latest 500
records in memory for that session.

#### How long history is kept

Rows older than \`telemetry.retention_days\` (default 7) are deleted at startup
and then every hour.

The same setting decides how long aiport keeps decided Suggestions and Change
requests, Catalog changes, and resolved or dismissed Issues. Dismissed ones
are kept at least 7 days, the quiet period. Open Issues are kept until they
resolve, and applied Changes are never deleted.

Deleted rows can linger in SQLite's files, because deletion is not secure
erasure.

#### Request ids

Every response carries \`X-Aiport-Request-Id\`. If an application logs it, you
can find the exact call in Activity. Votes refer to it too.

#### What was asked and what served it

For each model call, Activity keeps both sides: what the Application asked
for, and what served it.

The first side is the name it sent (a Level, a task name or a model id), its
module, and its effort along with where that effort came from: the
\`X-Aiport-Effort\` header, the provider's effort field, or neither, which
means \`medium\`. The second side is the Connection and model aiport sent it
to.

The Model column shows the name sent with the result beneath it, for example
\`balanced\` with \`chat/completions → small-model · low\` under it. **Details**
says it in a sentence, for example "Asked for balanced at low effort (from
the X-Aiport-Effort header); served by hosted with small-model."

The **Name** filter shows only calls that sent one name. Use it to check that
a Level or a per-effort mapping works, and to compare the models behind it.

A refused call has no serving model. Older records, from before this was
kept, show none of it.

#### Why a request failed

For a failed request, Activity shows the reason next to the status, for
example \`503 · Refused: no Connection for the Anthropic protocol\`.
**Details** shows the code, what it means, what to do, and a link to the page
that fixes it.

aiport keeps only short codes for this: its own refusal code, a connection
error class, and the provider's error type (such as \`rate_limit_error\`) when
the provider sends one. It never keeps the provider's error message, the
request or the response.

All reasons and their fixes are in [Activity reasons and their
fixes](#activity-reasons-and-their-fixes).

### Changes

Changes is where aiport adapts to the Applications that use it, always with
your approval. It holds open Issues, what waits for your decision, and every
configuration Change made through aiport, with Revert. It stays quiet: a
badge appears in the Monitor menu only when something needs a decision.

The page has five panels, from top to bottom:

- **Issues** holds the open Issues with their evidence, Explain, Fix it, and
  the fix on each side. See [Issues](#issues).
- **Suggestions and Change requests** is what waits for your decision, with
  the **Waiting / Accepted / Rejected** filter. See [What waits for your
  decision](#what-waits-for-your-decision).
- **Asked for, not supported yet** lists kinds of Change that Applications
  asked for and aiport cannot apply, with **Suggest to aiport**. See [Suggest
  to aiport and Share upstream](#suggest-to-aiport-and-share-upstream).
- **Extensions** shows the Extensions listed in \`aiport.yaml\` (your own
  programs that judge Change requests), with **Share upstream**. See [Add
  Extensions](#add-extensions).
- **Applied changes** lists every Change, newest first, with **Revert**, the
  automatic fixes switch (**Fix Issues automatically when the fix is clear**)
  and the sound. See [Applied changes and
  Revert](#applied-changes-and-revert).

The badge counts an Issue only while its fix waits for you. It also counts
the commands a Fix it run is waiting for you to approve.

### The Integrate page and the two briefs

Integrate is where you connect an application to aiport: it holds your
Applications, the onboarding wizard, and the instructions aiport writes for
the coding agents that do the work. The **Instructions** panel is at the top
(collapsed unless you have no Applications), with the Applications list and
the wizard below it.

aiport writes the two briefs from the running gateway, with its real address
and the protocols it has configured, and serves them at \`/_connect\`. The
**Application brief** is for the coding agent wiring the application to
aiport. The **Coding agent brief** is for coding agents that work in that
application's repository later.

Together with the communication standard, they are everything a coding agent
needs.

#### Applications

The Applications list shows every Application (an app or agent that calls
aiport with its own token), what it does through aiport, and what to do next.
Pairing requests waiting for your approval are always at the top.

Each pairing request shows its code, **Approve** (where you can also bind a
Connection and allow change requests) and **Reject**. See [Applications and
pairing](#applications-and-pairing).

From left to right, the columns show the name and client id, the
**Connection** (or the protocol default), **Changes** with the **May request
changes** toggle, **Last call** in the last 7 days (when activity recording
is on), **Issues**, **Guidance**, **Agent**, **Onboarding** progress such as
\`3/6\`, and finally **New token** and the trash button.

**Agent** names the Application's coding agent once it has one and a
repository path, and shows **Connect** otherwise. Select it to change them in
step 6 of the wizard.

**Issues** shows how many open or resolving Issues the Application has, with
a short cause such as "Unknown model llama3:8b" (or None). Select it to list
them under the row. Each Issue there has its summary, **Explain**, **Open in
Changes** (which opens it on the Changes page, where you approve or dismiss
it), and **Copy fix for the Coding agent**.

**Guidance** shows whether the Application's coding agent follows aiport's
current guidance (see [Guidance versions](#guidance-versions)). It reads
**Up to date**, **Behind v1 → v4** when the agent confirmed an older version
(the first number is the version it confirmed, the second the current one),
or **Unknown** when it never confirmed one. While it is not up to date, the
copy button next to it copies the refresh instruction, one paragraph to paste
to that coding agent.

Two more things can appear under Guidance. "Sends Levels" is a hint that the
Application's code may already use Levels, because its calls sent Levels in
the last 6 hours. It never counts as confirming. The **Modules** and **Votes**
chips show whether its calls in the last 7 days named their module, and
whether its users' votes arrive. When one is missing, hand its coding agent
the communication standard.

Under the Application's name, **Catalog** shows every name the Application
can send, with the Connection and model that serve it (as on the Routing
map), the models its Connections list, and its recent Catalog changes, newest
first. These are the same changes the Application reads at
\`/v1/aiport/catalog\`. **Levels**, next to Catalog, shows its Levels grid
(\`fast\`, \`balanced\`, \`best\`), with its own mappings marked **Own**. See [Map
Levels in the dashboard](#map-levels-in-the-dashboard).

Click the name itself to open **Issue new token** and **Remove application**
under its row. See [Give a token yourself, or remove an
Application](#give-a-token-yourself-or-remove-an-application).

Above the list, the **Approved / Rejected / All** filter picks what the list
shows. Rejected Applications stay hidden until you choose **Rejected** or
**All**, and each then shows when it was rejected, with **Forget**.
**Onboard a new application** makes the wizard follow a new one, and
selecting an existing Application makes it follow that one instead. **Add
with a token** adds an Application without pairing.

#### The instructions

You do not have to write instructions for the coding agent doing the work.
aiport writes them itself, from this gateway's real address and configured
protocols, and serves them at \`/_connect\`. The **Instructions** panel holds
the communication standard card and the two brief tiles.

The **Communication standard** card (\`/_connect?format=standard\`) shows the
prompt that hands the standard to a coding agent, with **Copy instruction**,
**Copy URL only** and **Read the standard**. See [The communication
standard](#the-communication-standard).

The **Application brief** (\`/_connect?format=handover&for=application\`) is
for the coding agent changing an application so its LLM calls go through
aiport. Its **Python / TypeScript / curl** picker switches the SDK snippets.
It covers:

- pairing
- keeping the token in \`.aiport\` (\`AIPORT_BASE_URL\`, \`AIPORT_APP_TOKEN\`), and
  what to do when it is lost
- pointing the OpenAI or Anthropic SDK at aiport
- removing provider keys from the application
- model names: send a Level or a task name, never a provider's model id, read
  \`GET /v1/models\` for the names, and \`level_unmapped\`
- saying how hard a call is with \`X-Aiport-Effort\`
- naming the module with \`X-Aiport-Module\`, and sending votes
- watching the \`X-Aiport-Catalog\` header
- errors and retries
- logging \`X-Aiport-Request-Id\`
- Issues: what they are, that the coding agent reads them with
  \`aiport_my_issues\`, that a Change that takes something away opens a
  \`catalog_changed\` one, and that aiport may fix routing itself as a
  revertable Change
- how to verify

The **Coding agent brief** (\`/_connect?format=handover&for=agent\`) is for the
coding agents that work in that application's repository later. It covers
the project MCP entry and the line for the repository's \`AGENTS.md\`, keeping
the guidance current, and the MCP tools with their HTTP equivalents. It also
tells the agent to check the Catalog, Issues and guidance at the start of a
session, to say when an Issue is fixed, and to name the need (a Level or task
name, and an effort) rather than a provider. The rest is about writing a good
Change request and waiting for your decision, the permission, and hints.

Each tile shows who the brief is for, **Copy whole brief** (the brief as
markdown, to paste to the coding agent) and its address. Open a tile to see
its steps and snippets, each with its own **Copy**. Snippets for a protocol
this gateway has not configured are marked "not configured".

To fetch the briefs without the dashboard, \`/_connect?format=handover\`
returns both, one after the other, and \`/_connect?format=markdown\` holds
everything, starting with **Start here**.

### Connections and keys

A Connection is a provider aiport can send calls to, and Providers is the page
where you add Connections, paste their keys and check they work. Applications
never see provider keys: they send their own aiport token, and aiport adds the
provider key on the way out.

aiport supports OpenAI, Anthropic, Amazon Bedrock, Meta Model API, and local
servers such as Ollama or llama.cpp. Each Connection speaks one protocol,
\`openai\` or \`anthropic\`. Each provider shows its logo on its card, in the add
list and in the Overview's provider distribution. llama.cpp and custom
providers show a neutral icon.

Connections are grouped as **Remote APIs** and **Local**. Both groups start
collapsed. Each heading shows how many Connections it holds and which need a
key or are unhealthy, and clicking it opens the group.

Use **Add a connection** to add one from a preset, or any other server that
speaks the OpenAI or Anthropic API as is (remote URLs must use HTTPS).
**Remove** takes one away, as described below. Both are Changes you can
revert in Changes.

To give a remote Connection its key, paste the provider API key into its
card. The key applies at once, is stored only in the private
\`aiport.yaml.env\` (never in \`aiport.yaml\`), and is never shown again.

You can also set the key in the shell aiport starts from, under the
Connection's key variable (see [configuration](configuration.md)). Provider
keys are never asked for in the terminal. Local Connections need no key.

**Check connection** (or **Check all** for a group) sends one authenticated
\`GET /models\` to the provider. It generates nothing and costs nothing, and it
shows whether the provider is reachable and accepted the key. A failed check
says why and what to do. The reason is either a connection error (the host
name does not resolve, the connection is refused, TLS failed, or it timed
out) or the HTTP status with the provider's own error type, for example
\`HTTP 401, authentication_error\`.

A passing check does not prove every model works.

**Show models** lists the models that Connection serves, from the same free
\`GET /models\` listing (OpenAI-, Anthropic- and Ollama-shaped answers). After a
check, the button shows the count ("Show 12 models").

aiport keeps only the model ids, in memory, and reuses a list for two
minutes. **Refresh** lists again, and so does **Check connection**. If the
listing fails (no key, key rejected, no model list, unreachable), aiport
treats that Connection's models as unknown, not as "serves nothing".
Disabled Connections are not listed.

**Make OpenAI default** / **Make Anthropic default** sets the Connection a
protocol uses when nothing else picks one.

Provider presets, endpoints and key variables are in
[configuration](configuration.md).

#### Removing a Connection, and when one fails

You can always remove a Connection: aiport moves everything that used it to
another Connection in the same step.

1. Select **Remove** on the Connection's card. A dialog lists everything still
   using it: a protocol default, an Application bound to it, an Alias (shared
   or an Application's own), Explanations, the model calls of Agent runs, and
   the compression evaluator.
2. Check the Connection picker on each item. It is already set to the best
   match:
   1. a Connection of the same protocol that lists the same model (for a
      binding or a protocol default, the one that lists most of the removed
      Connection's models)
   2. otherwise the protocol default (the new one, if you are removing the
      default)
   3. otherwise none.
3. Check the list of affected Applications below (see next).
4. Confirm.

Choosing **None** leaves an item without a Connection on purpose. For a
protocol default, that means the default is unset. A removed binding sends
the Application back to the protocol defaults, and a removed Alias retires
its task name. Explanations, Agent runs' model calls or the evaluator are
turned off (Agent runs then use the coding agent's own login).

Below the list, the dialog names the Applications whose Catalog (the names
they can send) the removal changes: task names removed or moved, or a binding
or protocol default they follow. It updates as you change a picker. aiport
itself is among them when the removal moves what its own model calls
(Explanations and Agent runs) use, which is \`agent_runs.connection\`, or the
protocol defaults when that is not set. aiport's Catalog changes are kept like
any Application's, but it gets no Issue, because it has no coding agent to
tell.

Confirming applies all the moves and the removal as **one Change**, so one
revert in Changes brings back the Connection and everything that used it.
Each affected Application is told through its Catalog. An Application that
lost something it used gets a Catalog-change Issue (see [A Catalog
change](#a-catalog-change)).

There is no automatic fallback when a Connection fails. If it is down or
keeps answering with errors, calls to it fail and the Application gets the
provider's own error. When this keeps happening for one Application, aiport
opens an Issue that offers to bind that Application to a working Connection,
and you approve it (see [A failing Connection](#a-failing-connection)).
aiport never moves an Application to another Connection by itself, so calls
never go where you did not choose.

Over the API, a \`DELETE /api/v1/targets/{id}\` with no plan (which is what you
send without the dialog) is refused when something uses the Connection,
Explanations and Agent runs' model calls included. The message names what
uses it. The dialog itself reads \`GET /api/v1/targets/{id}/dependents\` and
sends the plan as \`{"plan": {"<item>": "<Connection id or empty>"}}\`.

### Routing and protocol defaults

Routing decides which Connection and model serve each call, and the Routing
page is where you set the protocol defaults and the Levels, and see where
every name goes. aiport speaks each provider's own API and does not translate
between them: an OpenAI SDK call needs a Connection that speaks the \`openai\`
protocol, and an Anthropic SDK call needs one that speaks \`anthropic\`.

At the top of the page, **Protocol defaults** sets the Connection each
protocol uses when nothing else picks one. The **Levels** panel below it says
what \`fast\`, \`balanced\` and \`best\` mean here, and the **Routing map** shows
where every name goes.

Further down, **Connect an application** gives the two base URLs Applications
use (\`http://127.0.0.1:8787/v1\` for the OpenAI protocol,
\`http://127.0.0.1:8787\` for the Anthropic protocol), the \`/_connect\` protocol
description, and the telemetry queue. The **Applications** table and a
pointer to the instructions on Integrate come last.

Model allowlists are planned, not built.

#### How aiport picks a Connection and model

For each call, aiport goes down this list and uses the first rule that
applies:

1. A module route, when the Application has one for the module the call
   names (\`X-Aiport-Module\`) and the name it sends. See [Apply a Suggestion
   as a module route](#apply-a-suggestion-as-a-module-route).
2. A Level (\`fast\`, \`balanced\`, \`best\`), using the Application's own mapping
   on this protocol, else the shared one. A Level mapped on neither is
   refused with \`400 level_unmapped\`.
3. The Application's own Alias (a task name it alone uses) for the requested
   name, if it has one. It may name another Connection.
4. A shared task name marked \`everywhere\`, on its own Connection.
5. The Connection the Application is bound to, if any, or else the protocol
   default set under Protocol defaults, together with the shared Aliases of
   that Connection.

If the Level or Alias has per-effort models, aiport then picks the model for
the call's effort (see [Effort and reasoning](#effort-and-reasoning)). Only
the \`model\` field is replaced, plus the reasoning effort in the cases that
section describes.

When no Connection can serve the call:

| Situation | Answer |
| --- | --- |
| The protocol has no Connection | \`503 target_unconfigured\` |
| The Connection is disabled | \`503 target_disabled\` |
| The Connection speaks the other protocol | \`400 incompatible_protocol\` |

#### The Routing map

The Routing map answers "if this Application sends this name, which Connection
and model serve it?" for every Application at once. It is built from each
Application's Catalog, the same list its \`GET /v1/models\` answers, so the two
never disagree.

For each Application, the map lists every name it can send:

- its own task names
- the shared task names of the Connections it reaches, and the \`everywhere\`
  ones
- the Levels mapped for it (its own mapping or the shared one, per protocol)
- a row for **Any other model name**, which goes unchanged to its binding or
  to the protocol default (with the number of models that Connection lists)

Each row shows the protocol, the Connection with its provider's logo, and the
model aiport sends to the provider, with its per-effort models under it when
it has any. For example, docs-bot sending \`fast\` on the OpenAI protocol
reaches the local Connection with \`llama3:8b\`.

Choose an Application or a provider to narrow the map. The same view for one
Application, with its recent Catalog changes, is under **Catalog** on its row
on Integrate.

The pencil on a task name's row edits its per-effort models (see [Per-effort
models of a task name](#per-effort-models-of-a-task-name)). A Level's
per-effort models are edited in the Levels panel instead.

Scripts can read the map with the management token at \`GET /api/v1/routing\`,
adding \`?application=<id>\` for one Application with its \`changes\`. It holds
Application ids, names, task names, model names and Connection ids only,
never keys or addresses.

### Agents

The Agents page lists the coding agents whose own model calls go through
aiport, each with its own token, so they never see your provider keys. For
each, it shows the Application with its Connection (or the protocol default)
and the line that launches it.

To set one up and launch it:

1. Create the Application with \`aiport init --agent codex,claude\`. The agent
   needs an Application whose id is \`codex\` or \`claude\`. When none exists,
   the page tells you to run this.
2. In a terminal, run \`aiport agent codex --model MODEL\` or
   \`aiport agent claude --model MODEL\`.

Add \`--dry-run\` to see what would start, and put options for the agent after
\`--\`, as in \`-- agent-options\`. The settings apply to that launch only.
Credentials are never displayed.

This is separate from the Agent runs that [Connecting a coding
agent](#connecting-a-coding-agent) describes, where aiport runs a coding
agent on an Issue.

### Compression

The Compression page shows whether opt-in request compression saves anything
and whether it makes answers worse, for each combination of methods.
Compression only runs on requests that ask for it with \`X-Aiport-Compress\`
(see [Compress requests](#compress-requests)).

The page shows the savings and a trend of compressed against uncompressed
traffic (feedback, error rate, size saved). Token savings are estimated from
bytes, divided by 4.

Each method combination gets a verdict:

| Verdict | When |
| --- | --- |
| **Not enough data** | Fewer than 20 requests on either side. |
| **Possible harm** | Feedback more than 0.1 lower, or an error rate more than 5 points higher, than uncompressed traffic. |
| **No harm seen** | Otherwise. |

The **Evaluator** panel picks a Connection and model for model-based checks.
**Test evaluator** runs two known cases and reports whether the model labels
them correctly. A Connection in use as the evaluator cannot be removed until
you choose another one (or none), except through the removal dialog.

Details are in [compression](compression.md) and [compression
evaluators](compression-evaluators.md).

### Feedback

The Feedback page shows how the people using each application rate its
answers, per module, model and effort, so you can see which part of which
application does well and move the part that does not. Applications send the
votes, and [Collect feedback and improve
routing](#collect-feedback-and-improve-routing) explains how.

Pick the last 7, 30 or 90 days (30 by default).

For each Application, the page shows its thumbs up and down, its share of
up-votes, and how many of its calls were rated.

Its table lists each module, a named part of the application sent as
\`X-Aiport-Module\`. Inside each module there is one row per combination of
name sent (a Level, task name or model id), model and effort, with calls,
votes and the share of up-votes as a bar. The bar is green from 75%, amber
from 50% and red below 50%.

A module or combination with fewer than 5 votes is marked "few votes", which
means too few to judge. Calls that named no module are grouped as "No module
named". Guidance version 4 asks the application's coding agent to add
\`X-Aiport-Module\`.

With two or more judged modules, the page names the best and the weakest
module. Try another model or effort for the weakest under Routing, and watch
its share.

Under **Suggestions from the votes**, aiport lists what it found in the
votes. Each Suggestion is marked **Better elsewhere**, **One Level up** or
**Cheaper may do**, shows the move (from which model to which) and the
reason, and has an **Apply** button. See [Suggestions from the
votes](#suggestions-from-the-votes).

Each Application's panel also lists its **Module routes**, each with
**Remove**.

When there is nothing to show, the page says why. With request history off,
it says no calls or votes are kept. With calls but no votes, it says
applications start sending them once their coding agent follows guidance
version 4.

The same report is available to scripts as \`GET /api/v1/feedback?hours=N\` (1
to 2160, default 720), with \`applications\`, \`suggestions\` and
\`module_routes\`.

### Tasks

Tasks is a queue of work for coding agents: you write down tasks and decide
their intent and priority, and coding agents pick them up by name. Each task
gets a memorable name, such as \`task-amber-fox\`, that you can hand to a coding
agent. Click the name to copy it.

A task has a title of up to 200 characters, instructions, acceptance
criteria, a priority (None, Low, Medium or High) and a status: pending, in
progress, blocked, completed, failed or cancelled. Its history keeps who
changed what, and when.

By default, the filters above the queue show the open tasks (pending, in
progress or blocked). They can also show all of them, or one status. The
chime in the toolbar plays when no pending or in-progress tasks remain.

The queue is stored next to the configuration, in \`aiport.yaml.tasks.json\`.
From a terminal (\`--priority\` takes 0 to 3, None to High):

\`\`\`sh
aiport task list [--all]
aiport task show NAME
aiport task set NAME --status completed --priority 2 --note "tests pass"
\`\`\`

### Settings

Settings holds everything about this aiport install that is not routing:
sign-in, the default coding agent for Agent runs, who writes Issue
explanations, and the configuration file.

Under **Authentication** you set the username and password you sign in with,
and whether authentication is on. If you forget them, use **Forgot username
or password?** on the sign-in screen with the management token, then set a
new password here.

The **Default coding agent** is the coding agent of every Application that
has none of its own. The same panel holds **Model calls of Agent runs**, the
Anthropic-protocol Connection (and optional model) that those runs' calls go
through. See [Connecting a coding agent](#connecting-a-coding-agent) and
[Explain by a Coding agent](#explain-by-a-coding-agent).

In **Explanations** you choose the Connection and model that explain an Issue
when no coding agent does. See [Explain](#explain).

**Configuration** is about one file, \`aiport.yaml\`. Providers, routing, Levels
and per-effort models and each Application's coding agent are managed from
their pages. Task names arrive as approved Changes, or are edited in the file
along with the listen address. Validate hand edits with \`aiport check\`, then
restart.

## Reference

Lookup pages: the exact commands, endpoints, headers, error codes and settings,
for when you already know what you are looking for.

### Command line

These are the \`aiport\` commands you type in a terminal. Every command takes
\`--config PATH\` to use a different configuration file (default \`aiport.yaml\`).

| Command | What it does |
| --- | --- |
| \`aiport init\` | First-run setup. Writes \`aiport.yaml\` and \`aiport.yaml.env\`, and never overwrites them. Flags below. |
| \`aiport check\` | Checks the configuration. Lists Connections with no API key and Applications with no token, and warns once for each Alias named like a Level. |
| \`aiport serve\` | Starts aiport. Prints the dashboard address and, in a terminal, the management token. |
| \`aiport auth token\` | Prints the management token again. |
| \`aiport agent codex\`, \`aiport agent claude\` | Starts that coding agent through aiport, with its own token. Needs \`--model MODEL\`. |
| \`aiport task list\`, \`show NAME\`, \`set NAME\` | Reads and changes the task queue. |
| \`aiport token\` | Prints a new random token, for example for a token variable you set yourself. |
| \`aiport version\` | Prints the version. |

\`aiport init\` asks you to choose providers, coding agents and Levels (the
names \`fast\`, \`balanced\`, \`best\` that apps use instead of a model name). You
can answer with flags instead:

- \`--provider\`: comma-separated, from \`openai\`, \`anthropic\`, \`bedrock\`,
  \`meta\`, \`ollama\`, \`llamacpp\`.
- \`--agent codex,claude\`, or \`--agent none\`.
- \`--listen HOST:PORT\` (default \`127.0.0.1:8787\`).
- \`--region\` and \`--base-url\`.
- \`--levels LEVEL=CONNECTION/MODEL,...\`, or \`--no-levels\` to skip them.

A few other commands have flags of their own. For \`aiport agent\`, \`--dry-run\`
shows what would start without starting it, and anything after \`--\` goes to
the agent. \`aiport task list --all\` includes closed tasks, and \`aiport task
set\` takes \`--status\`, \`--priority 0-3\`, \`--note\` and \`--actor\`.

### Endpoints applications use

These are the addresses an Application calls. They are all on aiport's
address, \`http://127.0.0.1:8787\` by default.

Every endpoint needs the Application's own token as
\`Authorization: Bearer <token>\`. Anthropic SDKs may send it as \`x-api-key\`
instead. Pairing and \`/_connect\` need no token.

| Endpoint | What it is |
| --- | --- |
| \`POST /v1/chat/completions\`, \`POST /v1/responses\` | Model calls in the OpenAI protocol, sent to the provider as they are. |
| \`POST /v1/messages\` | Model calls in the Anthropic protocol, sent to the provider as they are. |
| \`GET /v1/models\` | The Application's Catalog (the model names it may use), in the SDK's own list format. |
| \`GET /_connect\` | How to connect, as JSON. \`?format=markdown\` gives the guide, \`?format=standard\` the communication standard, \`?format=handover&for=application\` or \`&for=agent\` a brief. No token. |
| \`POST /v1/aiport/pair\` | Asks to pair. Send \`{name, description?, repository?}\` and you get a \`code\` and a \`poll\` path. No token. |
| \`GET /v1/aiport/pair/{secret}\` | The \`poll\` path. Its \`status\` is \`pending\`, \`approved\` (with the \`token\`, shown once), \`rejected\` or \`expired\`. No token. |
| \`GET /v1/aiport/capabilities\` | What the Application can use and ask for right now. |
| \`GET /v1/aiport/catalog\` | The Catalog and its recent changes. |
| \`POST /v1/aiport/feedback\` | A vote on an answer: \`{request_id, vote}\` or \`{request_id, score}\`, with an optional \`module\`. |
| \`GET /v1/aiport/guidance\`, \`POST /v1/aiport/guidance/done\` | Guidance changes since the version last confirmed, and confirming the current one. |
| \`POST /v1/aiport/requests\`, \`GET /v1/aiport/requests\` (or \`/{id}\`) | File a Change request (ask the Owner to change a setting), and read the answers. |
| \`GET /v1/aiport/issues\` (or \`/{id}\`), \`POST /v1/aiport/issues/{id}/fixed\` | The Application's Issues, and saying one is fixed. |
| \`POST /mcp\` | The MCP endpoint for the Application's coding agent. Streamable HTTP, JSON-RPC sent as \`application/json\`. |

### Owner API

This is the API the dashboard itself uses, under \`/api/v1/\`. Your own scripts
can call it too, to do anything the dashboard does.

Send the management token as \`Authorization: Bearer <token>\`. The dashboard
uses its sign-in session instead, and Application tokens never work here.

Every request that changes something must also send \`X-Aiport-CSRF: 1\`, and
so must the two requests that ask providers for their model lists
(\`.../models\` and \`/serving\`).

Every edit here that changes the configuration is recorded as one Change you
can revert.

| Endpoint | What it does |
| --- | --- |
| \`GET /api/v1/auth/status\` | Whether a password is set, sign-in is on, and this browser is signed in. |
| \`POST /api/v1/auth/login\`, \`POST /api/v1/auth/logout\` | Sign in with \`{username, password}\`, or sign out. |
| \`POST /api/v1/auth/recover\` | Starts a session with the management token. Used by **Forgot username or password?**. |
| \`POST /api/v1/auth/settings\` | Sets \`{username, password, disabled}\`. |
| \`GET /api/v1/health\` | aiport is up, and the state of request history. |
| \`GET /api/v1/config\` | The configuration as the dashboard sees it, with which keys are set (never the keys themselves). |
| \`GET /api/v1/requests\` (or \`/{id}\`), \`GET /api/v1/stats\` | Request history, one record, and the summary on Overview. Filters: \`hours\` (1 to 168), \`origin\`, \`target\`, \`provider\`, \`outcome\`, \`model\`, \`from_ms\`/\`to_ms\`, \`limit\` (1 to 100), \`cursor\`. |
| \`POST /api/v1/targets\` | Adds a Connection: \`{id, provider, ...}\`. |
| \`PATCH /api/v1/targets/{id}\`, \`DELETE /api/v1/targets/{id}\` | Edits a Connection's \`region\`, \`endpoint\` or \`protocol\`, or removes one. Send \`{"plan": {...}}\` with the delete to re-point what used it. |
| \`PUT /api/v1/targets/{id}/api-key\`, \`DELETE .../api-key\` | Saves a provider key \`{api_key}\` to the private credential file, or removes it. |
| \`POST /api/v1/targets/{id}/check\` | **Check connection**: one signed-in \`GET /models\` to the provider. |
| \`GET /api/v1/targets/{id}/models\` | **Show models**: the model ids the Connection lists. |
| \`GET /api/v1/targets/{id}/dependents\`, \`POST .../dependents\` | What uses a Connection, with suggested replacements. \`POST {plan}\` previews a plan. |
| \`GET /api/v1/serving?protocol=P&model=M\` | Which Connections of a protocol list a model. \`&refresh=1\` asks again. |
| \`PUT /api/v1/defaults\` | Sets the protocol defaults: \`{openai_target, anthropic_target}\`. |
| \`GET /api/v1/pairings\` | Pairing requests waiting. \`?status=rejected\` lists the remembered rejected ones. |
| \`POST /api/v1/pairings/{code}/approve\`, \`.../reject\` | Approves or rejects a pairing request. Approve optionally takes \`{id, target, may_request_changes, agent, repository, replace}\`. |
| \`DELETE /api/v1/pairings/rejected/{id}\` | **Forget** a rejected Application. |
| \`POST /api/v1/origins\` | Adds an Application with a token: \`{name, id?, target?, may_request_changes?, agent?, repository?}\`. |
| \`PATCH /api/v1/origins/{id}\` | Sets an Application's \`may_request_changes\`, \`agent\` or \`repository\`. |
| \`POST /api/v1/origins/{id}/token\` | Issues a new token for an Application. The old one stops working. |
| \`DELETE /api/v1/origins/{id}\` | Removes an Application. |
| \`GET /api/v1/routing\` | The Routing map. \`?application=<id>\` for one Application, with its \`changes\`. |
| \`GET /api/v1/levels\` | Every Level mapping, shared and per Application, and the Connections. |
| \`PUT /api/v1/levels/{protocol}/{level}\` | Maps a Level: \`{application?, connection, model, effort?}\`. |
| \`DELETE /api/v1/levels/{protocol}/{level}\` | Clears a mapping: \`{application?}\`. |
| \`PUT /api/v1/aliases/effort\` | Sets the per-effort models of any Alias: \`{application?, alias, connection?, effort}\`. |
| \`PUT /api/v1/task-names/{name}\` | Sets a shared task name: \`{connection, model, effort?, everywhere?, reasoning?}\`. |
| \`DELETE /api/v1/task-names/{name}\` | Removes a shared task name: \`{connection}\`. |
| \`PUT /api/v1/module-routes/{application}/{module}/{name}\` | Sets a module route: \`{connection, model, why?}\`. |
| \`DELETE /api/v1/module-routes/{application}/{module}/{name}\` | Removes a module route. |
| \`GET /api/v1/feedback?hours=N\` | The votes report, Suggestions from the votes, and module routes. |
| \`GET /api/v1/compression?hours=N\` | The compression report. 1 to 2160 hours, default 168. |
| \`PUT /api/v1/evaluation\`, \`POST /api/v1/evaluation/test\` | Chooses the compression evaluator \`{connection, model}\`, and runs **Test evaluator**. |
| \`GET /api/v1/guidance\` | Each Application's guidance state, with \`behind\` and \`never_confirmed\`. |
| \`GET /api/v1/changes\`, \`POST /api/v1/changes/{id}/revert\` | Every Change, newest first, and **Revert** for one. |
| \`GET /api/v1/decisions?status=S\`, \`POST /api/v1/decisions/{id}\` | Suggestions and Change requests (\`pending\`, \`accepted\` or \`rejected\`). Decide one with \`{action: accept, reject or dismiss, note?, ...}\`. |
| \`GET /api/v1/issues\` | Issues, filtered by \`?state=\` and \`?application=\`, with \`fix_by\` per Application. |
| \`POST /api/v1/issues\` | Reports a failed request to its Application's coding agent: \`{request_id, note?}\`. |
| \`POST /api/v1/issues/{id}\` | Acts on an Issue, for example \`{"action": "reject"}\`. |
| \`POST /api/v1/issues/{id}/explain\` | Streams an explanation as JSON lines. |
| \`PUT /api/v1/adapting/explainer\` | Chooses who writes explanations: \`{connection, model}\`. |
| \`PUT /api/v1/adapting/auto-fix\` | Turns automatic fixes on or off: \`{enabled}\`. |
| \`POST /api/v1/issues/{id}/fix\`, \`GET /api/v1/issues/{id}/fix\` | Starts **Fix it**, and reads its \`branch_fix\` and \`diff\`. |
| \`POST /api/v1/issues/{id}/fix/keep\`, \`.../fix/discard\` | Keeps or discards a Fix it diff. |
| \`GET /api/v1/issues/{id}/run\` | The Issue's last Agent run, its log and approvals waiting. |
| \`GET /api/v1/runs\`, \`POST /api/v1/runs/{id}/stop\` | The Agent runs going now, and stopping one. |
| \`GET /api/v1/approvals\`, \`POST /api/v1/approvals/{id}\` | Commands waiting for you. Answer one with \`{answer: allow_once, allow_run or deny}\`. |
| \`GET /api/v1/agents\`, \`PUT /api/v1/agents/calls\` | The coding agents found, and where Agent runs send their model calls. |
| \`PUT /api/v1/agents/default\` | The default coding agent: \`{agent}\` (\`claude\`, \`codex\`, \`gemini\` or empty). |
| \`GET /api/v1/upstream\`, \`POST /api/v1/upstream/issues\` | Drafts for **Suggest to aiport** and **Share upstream**, and filing one after you confirm. |
| \`GET /api/v1/tasks\`, \`POST /api/v1/tasks\` | The task queue, and adding a task. |
| \`GET /api/v1/tasks/{name}\`, \`PATCH /api/v1/tasks/{name}\` | One task, and changing it. \`X-Aiport-Actor\` names who made the change. |

### Headers

These are the extra HTTP headers aiport reads on requests and adds to
responses. aiport removes every \`X-Aiport-*\` request header before it sends the
call on to the provider.

| Header | Direction | What it means |
| --- | --- | --- |
| \`Authorization: Bearer <token>\` or \`x-api-key\` | Request | The Application's own token. aiport swaps it for the provider key before sending the call on. |
| \`X-Aiport-Module\` | Request | Which part of the app the call is for. Required on every model call. Lower-case \`a\`–\`z\`, \`0\`–\`9\`, \`.\`, \`_\`, \`-\`, at most 64 characters. |
| \`X-Aiport-Effort\` | Request | \`low\`, \`medium\` or \`high\`: how hard the call is. Picks a per-effort model. If the body sets no reasoning effort, aiport also sends this to the provider as the reasoning effort. |
| \`X-Aiport-Compress\` | Request | \`whitespace\`, \`json\`, \`dedupe\` or \`all\` (comma-separated), or \`off\`. |
| \`X-Aiport-CSRF: 1\` | Request | Required on Owner API requests that change something. |
| \`X-Aiport-Actor\` | Request | Task API only (\`/api/v1/tasks\`): a short name for who made the change, kept in the task's history. Default \`owner\`. |
| \`X-Aiport-Request-Id\` | Response | aiport's id for the call. Keep it with the answer, for votes and logs. |
| \`X-Aiport-Model\` | Response | On every call sent on: the model that actually answered, after a Level, task name, per-effort model or module route replaced the name sent. A vote for this request id counts for this model. |
| \`X-Aiport-Route\` | Response | On every call sent on: \`connection=…; name=…; effort=…; via=…\`. \`via\` is \`level\`, \`task_name\`, \`module_route\` or \`model\`. |
| \`X-Aiport-Catalog\` | Response | The version of the Application's Catalog (12 hex characters). |
| \`X-Aiport-Compression\` | Response | The compression methods used, the \`status\` (\`applied\`, \`no_gain\` or \`failed\`), and the size in bytes before and after. |
| \`X-Aiport-Hint\` | Response | One line of help: the open Issue a failing call belongs to, the Change request that could fix it, or what to do about a refused token. |
| \`X-Aiport-Connect\` | Response | On \`/_connect\`: the version of the connection description (now \`1\`). |

\`/_connect\` also lists \`X-Aiport-Task\`, \`X-Aiport-Session\` and
\`X-Aiport-Version\` on requests, and \`X-Aiport-Source\` on responses. Those four
are planned, not built yet.

### Error codes

When aiport refuses or fails a call, the body is
\`{"error": {"code", "message", "request_id", ...}}\`. Any other error body comes
from the provider, and aiport passes it through unchanged, with its status and
headers.

The status and code never depend on the optional \`hint\`.

#### When aiport refuses an Application's token

A refused token gets either \`401 unauthorized\` (on a model or Application
endpoint) or \`403 application_token_required\`. Both answers carry an
\`error.fix\` block, written for the application's coding agent, and an
\`X-Aiport-Hint\` header that says the same in one line:

\`\`\`json
{"error": {"code": "unauthorized", "message": "...", "request_id": "...",
  "fix": {"for": "the application's coding agent",
          "steps": ["Check that .aiport exists at the repository root ...", "If .aiport is missing ..., ask the Owner ... New token ..."],
          "retry": "Once .aiport has the token, retry the same call ... Claude Code reads the MCP token when it connects: run /mcp and reconnect aiport, or start a new session.",
          "docs": "http://127.0.0.1:8787/_connect?format=markdown#if-aiport-refuses-the-token"}}}
\`\`\`

The steps tell the coding agent to:

1. Check that \`.aiport\` has the \`AIPORT_APP_TOKEN=\` line. An
   \`AIPORT_APP_TOKEN\` set in the environment wins over the file, so unset an
   old one.
2. If the token is missing, lost or revoked, ask the Owner to click
   **New token** on Integrate (**Issue token** when there is none yet), and
   paste the copied lines into \`.aiport\`.
3. Not pair again unless the Owner asks, and not look for a token anywhere
   else.
4. Retry the same call. If the application reads \`.aiport\` only at startup,
   restart it first.

#### Codes for applications

These are the codes an Application can get back from aiport itself, with what
each one means.

| Status and code | Meaning |
| --- | --- |
| \`401 unauthorized\` | The Application token is missing or wrong. Follow \`error.fix\`. |
| \`403 application_token_required\` | This endpoint serves one Application, and no Application token was recognized. The management token does not count. Follow \`error.fix\`. |
| \`403 invalid_host\`, \`403 invalid_origin\`, \`403 cross_site_request\` | The \`Host\` is not the listen address or a configured hostname, or a browser sent a cross-origin or cross-site request. All are refused. |
| \`403 run_credential\` | An Agent run's credential was used for something other than its model calls. |
| \`404 unsupported_path\`, \`405 method_not_allowed\` | Not part of aiport's protocol, or the wrong HTTP method. |
| \`400 incompatible_protocol\` | The chosen Connection speaks the other protocol. aiport does not translate between them. |
| \`400 level_unmapped\` | The Level is not mapped on this protocol, neither for this Application nor shared. The message names both. |
| \`400 ambiguous_model\` | aiport needs to replace the model name, so the body must be one JSON object with one \`model\` field. |
| \`400 request_read_failed\` | aiport could not read the request body (the Application stopped sending). Retry. |
| \`400 unknown_compression\` | \`X-Aiport-Compress\` names a method aiport does not offer. |
| \`413 request_too_large\` | The body is over the size limit (\`proxy.max_request_bytes\`). |
| \`429 concurrency_limit\` | Too many requests at once (\`proxy.max_concurrent\`). Retry with backoff. |
| \`503 target_unconfigured\` | No Connection for this protocol or Application. |
| \`503 target_disabled\` | The chosen Connection is turned off. |
| \`503 target_invalid\` | The chosen Connection's settings (its base URL) cannot be used. |
| \`503 upstream_key_missing\` | The Owner has not set the provider key yet. |
| \`502 upstream_unavailable\` | aiport could not connect to the provider. |
| \`502 upstream_request_failed\` | aiport could not build the request to the Connection's address. |
| \`504 upstream_timeout\` | The provider did not answer in time. |
| \`499 client_cancelled\` | The Application hung up first. Recorded in Activity, never sent. |
| \`400 invalid_format\` | \`/_connect\` got a \`format\` other than \`json\`, \`markdown\`, \`handover\` or \`standard\`, or a \`for\` other than \`application\` or \`agent\`. |
| \`415 unsupported_media_type\`, \`400 unsupported_protocol_version\` | \`/mcp\` takes JSON-RPC as \`application/json\`, in an MCP protocol version aiport supports. |
| \`400 invalid_pairing\`, \`404 pairing_not_found\`, \`429 pairing_busy\` | Pairing needs a name of 1 to 128 characters, the request expired or never existed, or too many are waiting (retry after a minute). |
| \`400 invalid_feedback\`, \`404 request_not_found\` | The vote is malformed, or aiport has no kept request with that id for this Application. |
| \`403 change_requests_not_allowed\` | The Owner has not allowed this Application to request changes. |
| \`400 invalid_change_request\` | The request is incomplete or too large. \`error.problems\` lists every problem. |
| \`403 change_request_refused\` | A policy Extension refused it. \`error.reason\` says why. |
| \`429 change_requests_rate_limited\` | More than 20 requests in the last hour. Retry after \`Retry-After\`. |
| \`404 change_request_not_found\`, \`404 issue_not_found\`, \`409 issue_not_open\` | Not this Application's request or Issue, or the Issue is no longer open. |
| \`409 adapting_off\` | The adapting loop is turned off in the configuration. |
| \`500 change_request_failed\`, \`500 issue_update_failed\`, \`500 guidance_not_saved\` | aiport could not save the Change request, the Issue or the guidance confirmation. Retry. |

The full list is at \`http://127.0.0.1:8787/_connect?format=markdown#errors\`.

#### Codes from the Owner API

The Owner API answers with the same error body. These are the codes you are
most likely to see:

- \`401 unauthorized\`: sign in again.
- \`403 csrf_required\`: send \`X-Aiport-CSRF: 1\`.
- \`503 config_readonly\`: aiport cannot write its configuration or credential
  file.
- \`409 revert_conflict\` and \`409 already_reverted\`: from **Revert**.
- \`409 token_from_environment\`: the token is set in the environment, so aiport
  cannot replace it.
- A \`404\` whose code ends in \`_not_found\`: the id is unknown.
- A \`400\` whose code starts with \`invalid_\`: the message says what to send.

### Configuration

All of aiport's settings live in one file, \`aiport.yaml\`.
[configuration](configuration.md) explains every key, and
[aiport.example.yaml](../aiport.example.yaml) is a starting point to copy.

When you change something in the dashboard, aiport writes it into the file in
place, keeps your other keys and comments, and applies it at once. If you edit
the file by hand, run \`aiport check\` and then restart aiport.

The top-level keys:

| Key | What it sets |
| --- | --- |
| \`version\` | The configuration format version. |
| \`listen\`, \`hostnames\` | The loopback address aiport listens on, and extra \`*.localhost\` names for the browser. |
| \`auth\` | Dashboard sign-in: \`username\`, \`password_hash\` (set it from Settings, never by hand), \`disabled\`, and \`token_env\`, the variable that holds the management token. |
| \`targets\` | The Connections. |
| \`defaults\` | The protocol defaults. |
| \`aliases\` | Shared Level mappings and task names (\`effort\`, \`everywhere\`, \`reasoning\`). |
| \`origins\` | The Applications: \`token_env\`, \`target\`, \`aliases\`, \`may_request_changes\`, \`agent\`, \`repository\`, \`module_routes\`. |
| \`telemetry.enabled\`, \`telemetry.retention_days\` | Request history on or off, and how long it, decided items and Catalog changes are kept (default 7 days). |
| other \`telemetry\` keys | \`queue_capacity\`, \`batch_size\`, \`flush_interval_ms\`, \`shutdown_drain_ms\`, \`max_event_bytes\` tune the recorder. |
| \`adapting\` | \`disabled\`, \`auto_fix\`, \`explain\`, the thresholds \`unknown_model\`, \`failing_connection\`, \`level_unmapped\` and \`unknown_token\`, and \`interval_seconds\`. |
| \`agent_runs\` | \`default_agent\`, \`connection\`, \`model\`, \`explain_minutes\`, \`fix_minutes\`. |
| \`evaluation\` | The compression evaluator's Connection and model. |
| \`extensions\` | Policy and guardrail programs. |
| \`upstream.repository\` | Where **Suggest to aiport** and **Share upstream** open drafts. |
| \`proxy\` | Limits: \`max_concurrent\`, \`max_request_bytes\`, \`connect_timeout_ms\`, \`request_read_timeout_ms\`, \`response_header_timeout_ms\`, \`stream_idle_timeout_ms\`. |

### Data and privacy

aiport keeps a record of each call, such as which model answered and how long
it took. It never keeps what you asked or what the model answered. Everything
stays on your machine.

#### The files

aiport keeps its files next to the configuration file:

| File | What it holds |
| --- | --- |
| \`aiport.yaml\` | The configuration. Never holds a key or token. |
| \`aiport.yaml.env\` | The private credentials, readable only by you (mode 0600): provider keys you paste, Application tokens, the management token. |
| \`aiport.yaml.db\` | A private SQLite database (mode 0600): request history, Changes, decisions, Issues, Catalog changes, guidance confirmations and rejected pairings. |
| \`aiport.yaml.tasks.json\` | The task queue (mode 0600). |
| \`aiport.yaml.worktrees/\` | The working copies Agent runs use while they run. |

#### What aiport records about each call

For each call aiport records the time, the route, the Application, the model
name it sent, the effort and where that effort came from, the module, the
Connection, the model that answered, the status, how long it took, short error
codes, the compression sizes and the vote.

aiport never records prompts, responses, request or response headers, or the
provider's error messages. Keys and tokens are not stored there either. Tokens
live only in the credential file.

Issues, Catalog changes and explanations hold only ids, model names, counts,
request ids and fixed advice text.

#### How long it is kept

Request history, decided Suggestions and Change requests, Catalog changes,
and resolved or dismissed Issues are kept for \`telemetry.retention_days\`
(7 days by default). Open Issues are kept until they are resolved.

Applied Changes are kept longer than that, and so are Issues that hold a fix
you reverted or rejected.

Some things are never written to disk at all. Model lists stay in memory for
two minutes, and Agent run logs stay in memory while their Issue is open.

When aiport deletes something, it is not securely erased from the disk.

### Security

aiport is built to be safe to run on your own development machine. It is
still experimental. To report a vulnerability privately, see
[SECURITY.md](../SECURITY.md).

Only your machine can reach it. aiport listens on \`127.0.0.1:8787\` by
default, which other computers cannot reach. It only answers requests
addressed to its own listen address or to the \`*.localhost\` names you
configured, and it refuses browser requests that come from other sites.

Apps never see your provider keys. Each Application has its own aiport
token, and aiport swaps that token for the real provider key when it sends
the call on. Calls only ever go to the Connections you set up.

The dashboard needs a sign-in. The first time, you sign in with the
management token, then set a password of at least 12 characters. A browser
session lasts 12 hours, or until aiport restarts. Requests that change
something also need the \`X-Aiport-CSRF: 1\` header, so another web page cannot
make them for you.

Nothing changes without you. Every Change is listed and can be reverted.
Automatic fixes are off by default, and even when they are on they can only
change one Application's Alias.

Agent runs, where aiport starts a coding agent to explain or fix an Issue,
start only when you click. A run works in a separate copy of the repository,
with a minimal environment and no aiport tokens or provider keys, and each
command it wants to run waits for your answer. aiport never commits to your
branch, merges or pushes. Read [What aiport cannot
see](#explain-by-a-coding-agent) before you use them.

Extensions, the policy and guardrail programs you can add, are opt-in. One
runs only if you list it with an absolute path. It has a timeout, and it
refuses unless it explicitly allows.

## Troubleshooting

Find what you see below, then follow the fix. For any failed request, open it
in Activity and click **Details**: it names the reason and links to the page
that fixes it. The exact reasons are listed in
[Activity reasons and their fixes](#activity-reasons-and-their-fixes).

#### The dashboard does not open

Either aiport is not running, or you are using a different address. Check that
\`aiport serve\` is running, and open the exact address it prints, since aiport
refuses other host names and requests from other sites. To use a name instead
of the number, see [Use a name instead of
127.0.0.1](#use-a-name-instead-of-127001).

#### I forgot the password

On the sign-in screen, click **Forgot username or password?** and paste the
token that \`aiport auth token\` prints. Then set a new password in Settings →
Authentication.

#### An application gets \`401 unauthorized\` or \`403 application_token_required\`

Its token is missing, lost, revoked or not loaded. The error has an
\`error.fix\` block that tells the app's coding agent what to do:

1. Check the token in \`.aiport\`.
2. If it is not there or does not work, click **New token** on the app's row
   on Integrate, and paste the lines into \`.aiport\`.
3. Restart the application if it reads \`.aiport\` only at startup.

Do not pair the app again.

#### A \`401\` never shows up in Activity

aiport cannot tell which Application sent it, so it has nowhere to list it.
Fix it as above. If the application never had a token, pair it on Integrate.

#### The aiport MCP server does not connect in Claude Code

The MCP entry read an old or missing token when it connected. Fix \`.aiport\`,
then run \`/mcp\` and reconnect aiport, or start a new session. A
\`403 application_token_required\` means no Application token was recognized.

#### An SDK, script or coding agent cannot reach \`aiport.localhost\`

Its resolver does not know the name, the name resolves to \`::1\`, or a sandbox
proxy refuses it. Use \`http://127.0.0.1:8787\` instead, which is the address
aiport always hands out. For a script, you can also add
\`127.0.0.1  aiport.localhost\` to \`/etc/hosts\`.

#### Calls show as \`unknown\` in Activity

The application is not sending its own token. It may be using a shared or old
one. Pair it, or give it a token with **Add with a token**.

#### Calls fail with \`400 level_unmapped\`

The Level is not mapped on that protocol. Go to Routing → Levels and map it,
or give the Application its own mapping on its Integrate row. If it keeps
happening, aiport opens an Issue.

#### Calls fail with \`503 target_unconfigured\`

There is no Connection for that protocol. Go to Routing → Protocol defaults
and pick one, or bind the Application to a Connection. Remember that aiport
does not translate between protocols.

#### The Feedback page shows no votes

Either request history is off, the application's coding agent has not
followed guidance version 4, or nobody has voted yet.

If history is off, turn on \`telemetry.enabled\`. If the coding agent is
behind, check the **Modules** and **Votes** chips on Integrate and hand it
the communication standard. Otherwise, wait for the first votes.

#### A vote gets \`404 request_not_found\`

The request id is not this Application's, the call is no longer kept, or
history is off. Send the \`X-Aiport-Request-Id\` of that answer, soon after the
call.

#### No Suggestions come from the votes

aiport only suggests a change when the evidence is clear. It needs at least 10
votes on the same module, name, model and effort, and the votes must not sit
between 60% and 90% thumbs up. Wait for more votes.

#### An application gets \`403 change_requests_not_allowed\`

You have not allowed that Application to request changes. Allow it on
Integrate if you want its requests.

#### An application gets \`409 adapting_off\`

The adapting loop is turned off in the configuration. To bring it back,
remove \`adapting.disabled: true\` and restart.

#### Explain is greyed out

No Connection is chosen for explanations, or adapting is off. Go to Settings →
Explanations and choose a Connection and model.

#### Revert refuses with a conflict

The configuration changed after that Change. Revert the newer Changes first.

#### Activity is empty

Request history is off, or it is not being saved. Check that
\`telemetry.enabled: true\` is set, and that the banner does not say history is
kept for this session only.

### Activity reasons and their fixes

When a request fails, Activity's **Details** shows a reason in plain words.
Find that text in the left column to see what to do.

| Activity says | Code | What to do |
| --- | --- | --- |
| Refused: no Connection for the Anthropic (or OpenAI) protocol | \`503 target_unconfigured\` | Routing → Protocol defaults: pick a Connection for that protocol, or bind the Application to one on Integrate. aiport does not translate between protocols. |
| Refused: the Connection is disabled | \`503 target_disabled\` | Turn it on in Providers, or pick another default on Routing. |
| Refused: the Connection's settings are invalid | \`503 target_invalid\` | Correct its base URL in Providers. |
| Refused: the Connection does not speak the … protocol | \`400 incompatible_protocol\` | Routing → Protocol defaults: pick a Connection of that protocol. |
| Refused: the Level is not mapped for the … protocol | \`400 level_unmapped\` | Routing → Levels: map the Level on that protocol, or give the Application its own mapping on its Integrate row. If it keeps happening, aiport opens an Issue (see [An unmapped Level](#an-unmapped-level)). |
| Refused: no API key for the Connection | \`503 upstream_key_missing\` | Paste the key on the Connection's card in Providers. |
| Refused: too many requests at once | \`429 concurrency_limit\` | Retry. If it keeps happening, raise \`proxy.max_concurrent\`. |
| Refused: the request is too large | \`413 request_too_large\` | Send less, or raise \`proxy.max_request_bytes\`. |
| Refused: the request body could not be read | \`400 request_read_failed\` | The Application stopped sending. Retry. |
| Refused: unknown compression method | \`400 unknown_compression\` | Use a method listed on the Compression page. |
| Refused: the model name is ambiguous | \`400 ambiguous_model\` | Send one JSON object with one \`model\` field. |
| Gateway error: could not build the provider request | \`502 upstream_request_failed\` | Check the Connection's base URL in Providers. |
| Cannot reach the provider: its host name does not resolve | \`502 upstream_unavailable\`, \`dns_failure\` | Check the base URL in Providers, and the network. |
| Cannot reach the provider: connection refused | \`502 upstream_unavailable\`, \`connection_refused\` | Start the local server (Ollama, llama.cpp) or fix the port, then **Check connection**. |
| Cannot reach the provider: the secure connection (TLS) failed | \`502 upstream_unavailable\`, \`tls_error\` | Check the scheme (\`http\` or \`https\`) and host name in the base URL. |
| Cannot reach the provider: connecting timed out, or the connection was closed | \`502 upstream_unavailable\`, \`timeout\` or \`connection_reset\` | Check the network and the base URL, then retry. |
| Timed out: the provider did not answer in time | \`504 upstream_timeout\` | Retry. For slow local models, raise \`proxy.response_header_timeout_ms\`. |
| Provider error: the key was rejected | provider \`401\` or \`403\` (\`authentication_error\`, \`invalid_api_key\`, \`permission_error\`) | Replace the key in Providers, then **Check connection**. |
| Provider error: unknown model | provider \`404\` (\`not_found_error\`, \`model_not_found\`) | Use a model the provider lists (the Application sees them with \`GET /v1/models\`), or add an Alias. If it keeps happening, aiport opens an Issue in Changes that says which Connection serves the model. |
| Provider error: rate-limited | provider \`429\` (\`rate_limit_error\`) | Wait, or ask the provider for a higher limit. |
| Provider error: out of quota or credit | \`insufficient_quota\`, \`billing_error\` | Check your plan and billing with the provider. |
| Provider error: the request was not accepted | provider \`400\` (\`invalid_request_error\`) | Check the model name and parameters the Application sends. |
| Provider error: the request is too large | provider \`413\` (\`request_too_large\`) | Send a smaller request or a shorter context. |
| Provider error: the provider failed | provider \`5xx\` (\`overloaded_error\`, \`api_error\`) | Retry later, and check the provider's status page. |
| The response broke off | \`incomplete\` | Retry. For long answers, raise \`proxy.stream_idle_timeout_ms\`. |
| Cancelled by the Application | \`499 client_cancelled\` | Nothing to fix. |

## What's new

Newest first. The commit history has every detail.

**2026-11-04**

- Every call aiport sends on now says which model answered
  (\`X-Aiport-Model\`) and how aiport chose it (\`X-Aiport-Route\`).
- Activity details open right under their row, and you can report a failed
  request to the Application's coding agent as an Issue.
- A live gauge sits at the top of every page, with Listening, one light per
  Application, and a meter of calls as they pass.
- The Catalog marks task names that route everywhere (\`everywhere: true\`).
- This manual was rewritten, with a quickstart, concepts, guides, a tour of
  the dashboard, reference and troubleshooting.
- Votes and Suggestions now work per module. Apps name the module of each
  call with \`X-Aiport-Module\`, and send a thumbs up or down per answer to
  \`/v1/aiport/feedback\`. The new [Feedback](#feedback) page shows the votes
  per Application, module, model and effort, and turns them into Suggestions
  you apply as module routes.
- The communication standard is out, at \`/_connect?format=standard\` and as
  the MCP tool \`aiport_standard\`. It has every rule, plus a checklist.
  Guidance version 4 asks coding agents to follow it and to add the thumbs to
  the app's screens.
- Integrate shows **Modules** and **Votes** chips for each Application, in the
  Guidance column.
- Task names can apply everywhere and carry a fixed reasoning effort. They
  have a new Owner API, \`/api/v1/task-names\`.
- \`X-Aiport-Effort\` on its own now also sets the provider's reasoning effort,
  when the body sets none and aiport chose the model.
- When aiport refuses a token, it tells the coding agent what to do, with an
  \`error.fix\` block and an \`X-Aiport-Hint\` header.

**2026-11-03**

- Levels (\`fast\`, \`balanced\`, \`best\`) arrived. They are mapped per protocol
  and per Application and edited in the dashboard, \`init\` proposes a mapping,
  and an Issue opens for unmapped Levels.
- Effort picks the model, with per-effort models for Levels and task names.
  Activity records the name, effort and model of each call.
- The Catalog: \`GET /v1/models\` per Application, \`X-Aiport-Catalog\`,
  \`/v1/aiport/catalog\`, Issues when the Catalog changes, and removing a
  Connection with a plan to re-point what used it.
- The Routing map shows every Application's names and what serves them.
- Guidance versions let coding agents know when to refresh, and Integrate
  shows who is behind.
- Tokens live in \`.aiport\`. You can issue a token yourself, add an Application
  without pairing, or remove one. Pairing survives restarts and repeated
  requests.
- Applications and agents are given \`127.0.0.1\`, not the hostname.
- The module path is \`github.com/harelix/aiport\`, so \`go install\` works.

**2026-11-02**

- Issues from start to finish: unknown models and failing Connections,
  **Copy fix for the Coding agent**, automatic fixes if you allow them, and
  the Application's coding agent hearing about its Issues.
- You can explain an Issue with a Connection, or with a coding agent that
  reads the code, and run **Fix it** on a branch, with a diff and approval for
  each command.
- The Integrate page arrived, with the onboarding wizard and both briefs.
  Also new were pairing, Change requests, Extensions, **Suggest to aiport**
  and **Share upstream**, and the MCP endpoint.
`,Wx=/(`[^`]+`|\*\*[^*]+\*\*|\[[^\]]+\]\([^)]+\))/g;function In({text:o}){return a.jsx(a.Fragment,{children:o.split(Wx).map((c,u)=>{if(c.startsWith("`")&&c.endsWith("`")&&c.length>1)return a.jsx("code",{children:c.slice(1,-1)},u);if(c.startsWith("**")&&c.endsWith("**")&&c.length>3)return a.jsx("strong",{children:a.jsx(In,{text:c.slice(2,-2)})},u);const r=/^\[([^\]]+)\]\(([^)]+)\)$/.exec(c);if(r){const[,h,m]=r;return m.startsWith("#")?a.jsx("a",{href:m,children:h},u):/^https?:\/\/(127\.0\.0\.1|localhost)[:/]/.test(m)||m.startsWith("/")?a.jsx("a",{href:m,target:"_blank",rel:"noreferrer",children:h},u):a.jsxs("span",{children:[h," (",a.jsx("code",{children:m.startsWith("docs/")||m.includes("/")?m:"docs/"+m}),")"]},u)}return a.jsx($i.Fragment,{children:c},u)})})}const Vs=o=>o.toLowerCase().replace(/[`*]/g,"").replace(/[^a-z0-9 -]/g,"").trim().replace(/\s+/g,"-");function Px(o,c){const u=o.replace(/\r\n/g,`
`).split(`
`),r=u.findIndex(v=>/^#{1,4}\s/.test(v)&&Vs(v.replace(/^#+\s+/,""))===Vs(c));if(r<0)return"";const h=/^#+/.exec(u[r])[0].length,m=[];let g=!1;for(const v of u.slice(r+1)){v.startsWith("```")&&(g=!g);const b=!g&&/^(#{1,4})\s/.exec(v);if(b&&b[1].length<=h)break;m.push(v)}return m.join(`
`).trim()}function $x(o,c=2){const u=[];for(const r of o.split(/\n\s*\n/)){if(/^#{1,4}\s/.test(r.trim())||u.length>=c)break;r.trim()&&u.push(r.trim())}return u.join(`

`)}function ew(o,c=2){const u=[];let r=!1;for(const h of o.split(`
`)){h.startsWith("```")&&(r=!r);const m=!r&&/^(#{2,4})\s+(.*)$/.exec(h);m&&m[1].length-1<=c&&u.push([m[1].length-1,m[2]])}return u}function cv({source:o,anchors:c=!1}){const u=o.replace(/\r\n/g,`
`).split(`
`),r=[];let h=0;for(;h<u.length;){const m=u[h];if(m.startsWith("```")){const C=[];for(h++;h<u.length&&!u[h].startsWith("```");h++)C.push(u[h]);h++,r.push(a.jsx("pre",{className:"snippet-code",children:a.jsx("code",{children:C.join(`
`)})},r.length));continue}const g=/^(#{1,4})\s+(.*)$/.exec(m);if(g){const x="h"+Math.min(g[1].length+1,5),y=Vs(g[2]);r.push(a.jsxs(x,{id:y,children:[a.jsx(In,{text:g[2]}),c&&a.jsx("a",{className:"heading-anchor",href:"#"+y,"aria-label":"Link to "+g[2],onClick:w=>{w.preventDefault(),document.getElementById(y)?.scrollIntoView({block:"start"})},children:"#"})]},r.length)),h++;continue}if(/^\|/.test(m)){const C=[];for(;h<u.length&&/^\|/.test(u[h]);h++)/^\|[\s|:-]+\|$/.test(u[h])||C.push(u[h].replace(/^\||\|$/g,"").split("|").map(w=>w.trim()));const[x,...y]=C;r.push(a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsx("tr",{children:x.map((w,A)=>a.jsx("th",{children:a.jsx(In,{text:w})},A))})}),a.jsx("tbody",{children:y.map((w,A)=>a.jsx("tr",{children:w.map((E,_)=>a.jsx("td",{children:a.jsx(In,{text:E})},_))},A))})]})},r.length));continue}const v=/^(\s*)([-*]|\d+\.)\s+(.*)$/.exec(m);if(v){const C=/\d/.test(v[2]),x=[];for(;h<u.length;h++){const w=/^(\s*)([-*]|\d+\.)\s+(.*)$/.exec(u[h]);if(w)x.push(w[3]);else if(/^\s{2,}\S/.test(u[h])&&x.length)x[x.length-1]+=" "+u[h].trim();else break}const y=C?"ol":"ul";r.push(a.jsx(y,{children:x.map((w,A)=>a.jsx("li",{children:a.jsx(In,{text:w})},A))},r.length));continue}if(!m.trim()||m.trim()==="---"){h++;continue}const b=[];for(;h<u.length&&u[h].trim()&&!/^(#{1,4}\s|```|\||\s*([-*]|\d+\.)\s)/.test(u[h]);h++)b.push(u[h].trim());r.push(a.jsx("p",{children:a.jsx(In,{text:b.join(" ")})},r.length))}return a.jsx(a.Fragment,{children:r})}function tw(o,c=480){let u=$x(o,1);const r=/^\s*(?:[-*]|\d+\.)\s+/;if(r.test(u)){const g=u.split(`
`),v=[g[0].replace(r,"")];for(const b of g.slice(1)){if(r.test(b))break;v.push(b.trim())}u=v.join(" ")}if(u=u.replace(/\s*\n\s*/g," ").trim(),u.length<=c)return u;const h=u.slice(0,c),m=Math.max(h.lastIndexOf(". "),h.lastIndexOf("; "));return(m>c/3?h.slice(0,m+1):h.replace(/\s+\S*$/,""))+" …"}const dv=f.createContext(()=>{}),Nm={Overview:"A tour of the dashboard",Activity:"Activity and history",Changes:"Changes",Integrate:"The Integrate page and the two briefs",Providers:"Connections and keys",Routing:"Routing and protocol defaults",Agents:"Agents",Compression:"Compression",Feedback:"Feedback",Tasks:"Tasks",Settings:"Settings"};function wt({section:o}){const c=f.useContext(dv),[u,r]=f.useState(!1),[h,m]=f.useState({top:0,left:0}),g=f.useRef(null),v=f.useRef(null),b=f.useId(),C=tw(Px(_d,o));f.useEffect(()=>{if(!u)return;const y=A=>{(A instanceof KeyboardEvent?A.key==="Escape":!v.current?.contains(A.target)&&!g.current?.contains(A.target))&&r(!1)},w=A=>{v.current?.contains(A.target)||r(!1)};return document.addEventListener("keydown",y),document.addEventListener("mousedown",y),window.addEventListener("scroll",w,!0),()=>{document.removeEventListener("keydown",y),document.removeEventListener("mousedown",y),window.removeEventListener("scroll",w,!0)}},[u]),f.useLayoutEffect(()=>{if(!u||!v.current||!g.current)return;const y=g.current.getBoundingClientRect(),w=v.current.offsetHeight;y.bottom+8+w>window.innerHeight-12&&y.top-8-w>=12&&m(A=>({...A,top:y.top-8-w}))},[u]);function x(){const y=g.current.getBoundingClientRect();m({top:y.bottom+8,left:Math.max(12,Math.min(y.left-16,window.innerWidth-380-12))}),r(A=>!A)}return a.jsxs(a.Fragment,{children:[a.jsx("button",{ref:g,type:"button",className:"help-link"+(u?" on":""),onClick:x,"aria-expanded":u,"aria-controls":b,title:"What is this? "+o,"aria-label":"Explain: "+o,children:a.jsx(Sx,{size:14,strokeWidth:1.75,"aria-hidden":!0})}),u&&a.jsxs("div",{ref:v,id:b,className:"help-pop",role:"dialog","aria-label":o,style:{top:h.top,left:h.left},onClick:y=>y.stopPropagation(),children:[a.jsxs("div",{className:"help-pop-head",children:[a.jsx("strong",{children:o}),a.jsx("button",{type:"button",className:"icon-button",onClick:()=>r(!1),"aria-label":"Close",children:a.jsx(wa,{size:14,strokeWidth:1.75,"aria-hidden":!0})})]}),a.jsx("div",{className:"help-pop-body brief",children:C?a.jsx(cv,{source:C}):a.jsx("p",{children:"Read about this in the manual."})}),a.jsxs("button",{type:"button",className:"text-button help-pop-more",onClick:()=>{r(!1),c(o)},children:["Read the whole section in the manual ",a.jsx(dn,{size:14,"aria-hidden":!0})]})]})]})}const nw={openai:"M22.2819 9.8211a5.9847 5.9847 0 0 0-.5157-4.9108 6.0462 6.0462 0 0 0-6.5098-2.9A6.0651 6.0651 0 0 0 4.9807 4.1818a5.9847 5.9847 0 0 0-3.9977 2.9 6.0462 6.0462 0 0 0 .7427 7.0966 5.98 5.98 0 0 0 .511 4.9107 6.051 6.051 0 0 0 6.5146 2.9001A5.9847 5.9847 0 0 0 13.2599 24a6.0557 6.0557 0 0 0 5.7718-4.2058 5.9894 5.9894 0 0 0 3.9977-2.9001 6.0557 6.0557 0 0 0-.7475-7.0729zm-9.022 12.6081a4.4755 4.4755 0 0 1-2.8764-1.0408l.1419-.0804 4.7783-2.7582a.7948.7948 0 0 0 .3927-.6813v-6.7369l2.02 1.1686a.071.071 0 0 1 .038.052v5.5826a4.504 4.504 0 0 1-4.4945 4.4944zm-9.6607-4.1254a4.4708 4.4708 0 0 1-.5346-3.0137l.142.0852 4.783 2.7582a.7712.7712 0 0 0 .7806 0l5.8428-3.3685v2.3324a.0804.0804 0 0 1-.0332.0615L9.74 19.9502a4.4992 4.4992 0 0 1-6.1408-1.6464zM2.3408 7.8956a4.485 4.485 0 0 1 2.3655-1.9728V11.6a.7664.7664 0 0 0 .3879.6765l5.8144 3.3543-2.0201 1.1685a.0757.0757 0 0 1-.071 0l-4.8303-2.7865A4.504 4.504 0 0 1 2.3408 7.872zm16.5963 3.8558L13.1038 8.364 15.1192 7.2a.0757.0757 0 0 1 .071 0l4.8303 2.7913a4.4944 4.4944 0 0 1-.6765 8.1042v-5.6772a.79.79 0 0 0-.407-.667zm2.0107-3.0231l-.142-.0852-4.7735-2.7818a.7759.7759 0 0 0-.7854 0L9.409 9.2297V6.8974a.0662.0662 0 0 1 .0284-.0615l4.8303-2.7866a4.4992 4.4992 0 0 1 6.6802 4.66zM8.3065 12.863l-2.02-1.1638a.0804.0804 0 0 1-.038-.0567V6.0742a4.4992 4.4992 0 0 1 7.3757-3.4537l-.142.0805L8.704 5.459a.7948.7948 0 0 0-.3927.6813zm1.0976-2.3654l2.602-1.4998 2.6069 1.4998v2.9994l-2.5974 1.4997-2.6067-1.4997Z",anthropic:"M17.3041 3.541h-3.6718l6.696 16.918H24Zm-10.6082 0L0 20.459h3.7442l1.3693-3.5527h7.0052l1.3693 3.5528h3.7442L10.5363 3.5409Zm-.3712 10.2232 2.2914-5.9456 2.2914 5.9456Z",bedrock:"M6.763 10.036c0 .296.032.535.088.71.064.176.144.368.256.576.04.063.056.127.056.183 0 .08-.048.16-.152.24l-.503.335a.383.383 0 0 1-.208.072c-.08 0-.16-.04-.239-.112a2.47 2.47 0 0 1-.287-.375 6.18 6.18 0 0 1-.248-.471c-.622.734-1.405 1.101-2.347 1.101-.67 0-1.205-.191-1.596-.574-.391-.384-.59-.894-.59-1.533 0-.678.239-1.23.726-1.644.487-.415 1.133-.623 1.955-.623.272 0 .551.024.846.064.296.04.6.104.918.176v-.583c0-.607-.127-1.03-.375-1.277-.255-.248-.686-.367-1.3-.367-.28 0-.568.031-.863.103-.295.072-.583.16-.862.272a2.287 2.287 0 0 1-.28.104.488.488 0 0 1-.127.023c-.112 0-.168-.08-.168-.247v-.391c0-.128.016-.224.056-.28a.597.597 0 0 1 .224-.167c.279-.144.614-.264 1.005-.36a4.84 4.84 0 0 1 1.246-.151c.95 0 1.644.216 2.091.647.439.43.662 1.085.662 1.963v2.586zm-3.24 1.214c.263 0 .534-.048.822-.144.287-.096.543-.271.758-.51.128-.152.224-.32.272-.512.047-.191.08-.423.08-.694v-.335a6.66 6.66 0 0 0-.735-.136 6.02 6.02 0 0 0-.75-.048c-.535 0-.926.104-1.19.32-.263.215-.39.518-.39.917 0 .375.095.655.295.846.191.2.47.296.838.296zm6.41.862c-.144 0-.24-.024-.304-.08-.064-.048-.12-.16-.168-.311L7.586 5.55a1.398 1.398 0 0 1-.072-.32c0-.128.064-.2.191-.2h.783c.151 0 .255.025.31.08.065.048.113.16.16.312l1.342 5.284 1.245-5.284c.04-.16.088-.264.151-.312a.549.549 0 0 1 .32-.08h.638c.152 0 .256.025.32.08.063.048.12.16.151.312l1.261 5.348 1.381-5.348c.048-.16.104-.264.16-.312a.52.52 0 0 1 .311-.08h.743c.127 0 .2.065.2.2 0 .04-.009.08-.017.128a1.137 1.137 0 0 1-.056.2l-1.923 6.17c-.048.16-.104.263-.168.311a.51.51 0 0 1-.303.08h-.687c-.151 0-.255-.024-.32-.08-.063-.056-.119-.16-.15-.32l-1.238-5.148-1.23 5.14c-.04.16-.087.264-.15.32-.065.056-.177.08-.32.08zm10.256.215c-.415 0-.83-.048-1.229-.143-.399-.096-.71-.2-.918-.32-.128-.071-.215-.151-.247-.223a.563.563 0 0 1-.048-.224v-.407c0-.167.064-.247.183-.247.048 0 .096.008.144.024.048.016.12.048.2.08.271.12.566.215.878.279.319.064.63.096.95.096.502 0 .894-.088 1.165-.264a.86.86 0 0 0 .415-.758.777.777 0 0 0-.215-.559c-.144-.151-.416-.287-.807-.415l-1.157-.36c-.583-.183-1.014-.454-1.277-.813a1.902 1.902 0 0 1-.4-1.158c0-.335.073-.63.216-.886.144-.255.335-.479.575-.654.24-.184.51-.32.83-.415.32-.096.655-.136 1.006-.136.175 0 .359.008.535.032.183.024.35.056.518.088.16.04.312.08.455.127.144.048.256.096.336.144a.69.69 0 0 1 .24.2.43.43 0 0 1 .071.263v.375c0 .168-.064.256-.184.256a.83.83 0 0 1-.303-.096 3.652 3.652 0 0 0-1.532-.311c-.455 0-.815.071-1.062.223-.248.152-.375.383-.375.71 0 .224.08.416.24.567.159.152.454.304.877.44l1.134.358c.574.184.99.44 1.237.767.247.327.367.702.367 1.117 0 .343-.072.655-.207.926-.144.272-.336.511-.583.703-.248.2-.543.343-.886.447-.36.111-.734.167-1.142.167zM21.698 16.207c-2.626 1.94-6.442 2.969-9.722 2.969-4.598 0-8.74-1.7-11.87-4.526-.247-.223-.024-.527.272-.351 3.384 1.963 7.559 3.153 11.877 3.153 2.914 0 6.114-.607 9.06-1.852.439-.2.814.287.383.607zM22.792 14.961c-.336-.43-2.22-.207-3.074-.103-.255.032-.295-.192-.063-.36 1.5-1.053 3.967-.75 4.254-.399.287.36-.08 2.826-1.485 4.007-.215.184-.423.088-.327-.151.32-.79 1.03-2.57.695-2.994z",meta:"M6.915 4.03c-1.968 0-3.683 1.28-4.871 3.113C.704 9.208 0 11.883 0 14.449c0 .706.07 1.369.21 1.973a6.624 6.624 0 0 0 .265.86 5.297 5.297 0 0 0 .371.761c.696 1.159 1.818 1.927 3.593 1.927 1.497 0 2.633-.671 3.965-2.444.76-1.012 1.144-1.626 2.663-4.32l.756-1.339.186-.325c.061.1.121.196.183.3l2.152 3.595c.724 1.21 1.665 2.556 2.47 3.314 1.046.987 1.992 1.22 3.06 1.22 1.075 0 1.876-.355 2.455-.843a3.743 3.743 0 0 0 .81-.973c.542-.939.861-2.127.861-3.745 0-2.72-.681-5.357-2.084-7.45-1.282-1.912-2.957-2.93-4.716-2.93-1.047 0-2.088.467-3.053 1.308-.652.57-1.257 1.29-1.82 2.05-.69-.875-1.335-1.547-1.958-2.056-1.182-.966-2.315-1.303-3.454-1.303zm10.16 2.053c1.147 0 2.188.758 2.992 1.999 1.132 1.748 1.647 4.195 1.647 6.4 0 1.548-.368 2.9-1.839 2.9-.58 0-1.027-.23-1.664-1.004-.496-.601-1.343-1.878-2.832-4.358l-.617-1.028a44.908 44.908 0 0 0-1.255-1.98c.07-.109.141-.224.211-.327 1.12-1.667 2.118-2.602 3.358-2.602zm-10.201.553c1.265 0 2.058.791 2.675 1.446.307.327.737.871 1.234 1.579l-1.02 1.566c-.757 1.163-1.882 3.017-2.837 4.338-1.191 1.649-1.81 1.817-2.486 1.817-.524 0-1.038-.237-1.383-.794-.263-.426-.464-1.13-.464-2.046 0-2.221.63-4.535 1.66-6.088.454-.687.964-1.226 1.533-1.533a2.264 2.264 0 0 1 1.088-.285z",ollama:"M16.361 10.26a.894.894 0 0 0-.558.47l-.072.148.001.207c0 .193.004.217.059.353.076.193.152.312.291.448.24.238.51.3.872.205a.86.86 0 0 0 .517-.436.752.752 0 0 0 .08-.498c-.064-.453-.33-.782-.724-.897a1.06 1.06 0 0 0-.466 0zm-9.203.005c-.305.096-.533.32-.65.639a1.187 1.187 0 0 0-.06.52c.057.309.31.59.598.667.362.095.632.033.872-.205.14-.136.215-.255.291-.448.055-.136.059-.16.059-.353l.001-.207-.072-.148a.894.894 0 0 0-.565-.472 1.02 1.02 0 0 0-.474.007Zm4.184 2c-.131.071-.223.25-.195.383.031.143.157.288.353.407.105.063.112.072.117.136.004.038-.01.146-.029.243-.02.094-.036.194-.036.222.002.074.07.195.143.253.064.052.076.054.255.059.164.005.198.001.264-.03.169-.082.212-.234.15-.525-.052-.243-.042-.28.087-.355.137-.08.281-.219.324-.314a.365.365 0 0 0-.175-.48.394.394 0 0 0-.181-.033c-.126 0-.207.03-.355.124l-.085.053-.053-.032c-.219-.13-.259-.145-.391-.143a.396.396 0 0 0-.193.032zm.39-2.195c-.373.036-.475.05-.654.086-.291.06-.68.195-.951.328-.94.46-1.589 1.226-1.787 2.114-.04.176-.045.234-.045.53 0 .294.005.357.043.524.264 1.16 1.332 2.017 2.714 2.173.3.033 1.596.033 1.896 0 1.11-.125 2.064-.727 2.493-1.571.114-.226.169-.372.22-.602.039-.167.044-.23.044-.523 0-.297-.005-.355-.045-.531-.288-1.29-1.539-2.304-3.072-2.497a6.873 6.873 0 0 0-.855-.031zm.645.937a3.283 3.283 0 0 1 1.44.514c.223.148.537.458.671.662.166.251.26.508.303.82.02.143.01.251-.043.482-.08.345-.332.705-.672.957a3.115 3.115 0 0 1-.689.348c-.382.122-.632.144-1.525.138-.582-.006-.686-.01-.853-.042-.57-.107-1.022-.334-1.35-.68-.264-.28-.385-.535-.45-.946-.03-.192.025-.509.137-.776.136-.326.488-.73.836-.963.403-.269.934-.46 1.422-.512.187-.02.586-.02.773-.002zm-5.503-11a1.653 1.653 0 0 0-.683.298C5.617.74 5.173 1.666 4.985 2.819c-.07.436-.119 1.04-.119 1.503 0 .544.064 1.24.155 1.721.02.107.031.202.023.208a8.12 8.12 0 0 1-.187.152 5.324 5.324 0 0 0-.949 1.02 5.49 5.49 0 0 0-.94 2.339 6.625 6.625 0 0 0-.023 1.357c.091.78.325 1.438.727 2.04l.13.195-.037.064c-.269.452-.498 1.105-.605 1.732-.084.496-.095.629-.095 1.294 0 .67.009.803.088 1.266.095.555.288 1.143.503 1.534.071.128.243.393.264.407.007.003-.014.067-.046.141a7.405 7.405 0 0 0-.548 1.873c-.062.417-.071.552-.071.991 0 .56.031.832.148 1.279L3.42 24h1.478l-.05-.091c-.297-.552-.325-1.575-.068-2.597.117-.472.25-.819.498-1.296l.148-.29v-.177c0-.165-.003-.184-.057-.293a.915.915 0 0 0-.194-.25 1.74 1.74 0 0 1-.385-.543c-.424-.92-.506-2.286-.208-3.451.124-.486.329-.918.544-1.154a.787.787 0 0 0 .223-.531c0-.195-.07-.355-.224-.522a3.136 3.136 0 0 1-.817-1.729c-.14-.96.114-2.005.69-2.834.563-.814 1.353-1.336 2.237-1.475.199-.033.57-.028.776.01.226.04.367.028.512-.041.179-.085.268-.19.374-.431.093-.215.165-.333.36-.576.234-.29.46-.489.822-.729.413-.27.884-.467 1.352-.561.17-.035.25-.04.569-.04.319 0 .398.005.569.04a4.07 4.07 0 0 1 1.914.997c.117.109.398.457.488.602.034.057.095.177.132.267.105.241.195.346.374.43.14.068.286.082.503.045.343-.058.607-.053.943.016 1.144.23 2.14 1.173 2.581 2.437.385 1.108.276 2.267-.296 3.153-.097.15-.193.27-.333.419-.301.322-.301.722-.001 1.053.493.539.801 1.866.708 3.036-.062.772-.26 1.463-.533 1.854a2.096 2.096 0 0 1-.224.258.916.916 0 0 0-.194.25c-.054.109-.057.128-.057.293v.178l.148.29c.248.476.38.823.498 1.295.253 1.008.231 2.01-.059 2.581a.845.845 0 0 0-.044.098c0 .006.329.009.732.009h.73l.02-.074.036-.134c.019-.076.057-.3.088-.516.029-.217.029-1.016 0-1.258-.11-.875-.295-1.57-.597-2.226-.032-.074-.053-.138-.046-.141.008-.005.057-.074.108-.152.376-.569.607-1.284.724-2.228.031-.26.031-1.378 0-1.628-.083-.645-.182-1.082-.348-1.525a6.083 6.083 0 0 0-.329-.7l-.038-.064.131-.194c.402-.604.636-1.262.727-2.04a6.625 6.625 0 0 0-.024-1.358 5.512 5.512 0 0 0-.939-2.339 5.325 5.325 0 0 0-.95-1.02 8.097 8.097 0 0 1-.186-.152.692.692 0 0 1 .023-.208c.208-1.087.201-2.443-.017-3.503-.19-.924-.535-1.658-.98-2.082-.354-.338-.716-.482-1.15-.455-.996.059-1.8 1.205-2.116 3.01a6.805 6.805 0 0 0-.097.726c0 .036-.007.066-.015.066a.96.96 0 0 1-.149-.078A4.857 4.857 0 0 0 12 3.03c-.832 0-1.687.243-2.456.698a.958.958 0 0 1-.148.078c-.008 0-.015-.03-.015-.066a6.71 6.71 0 0 0-.097-.725C8.997 1.392 8.337.319 7.46.048a2.096 2.096 0 0 0-.585-.041Zm.293 1.402c.248.197.523.759.682 1.388.03.113.06.244.069.292.007.047.026.152.041.233.067.365.098.76.102 1.24l.002.475-.12.175-.118.178h-.278c-.324 0-.646.041-.954.124l-.238.06c-.033.007-.038-.003-.057-.144a8.438 8.438 0 0 1 .016-2.323c.124-.788.413-1.501.696-1.711.067-.05.079-.049.157.013zm9.825-.012c.17.126.358.46.498.888.28.854.36 2.028.212 3.145-.019.14-.024.151-.057.144l-.238-.06a3.693 3.693 0 0 0-.954-.124h-.278l-.119-.178-.119-.175.002-.474c.004-.669.066-1.19.214-1.772.157-.623.434-1.185.68-1.382.078-.062.09-.063.159-.012z"};function Ks({provider:o,location:c,size:u=18}){const r=nw[o];if(r)return a.jsx("svg",{className:"provider-logo",viewBox:"0 0 24 24",width:u,height:u,fill:"currentColor","aria-hidden":!0,focusable:"false",children:a.jsx("path",{d:r})});const h=o==="llamacpp"?Ax:c==="local"?Ix:mg;return a.jsx(h,{className:"provider-logo",size:u,strokeWidth:1.75,"aria-hidden":!0})}async function aw(o,c,u){const r=await fetch("/api/v1/upstream/issues",{method:"POST",credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:JSON.stringify({...o,confirm:!0,title:c,body:u})}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Request failed ("+r.status+").");return h.url}function sw(o,c,u){return"https://github.com/"+o+"/issues/new?"+new URLSearchParams({body:u,title:c}).toString()}function _m({label:o,draft:c,what:u,upstream:r}){const[h,m]=f.useState(!1),[g,v]=f.useState(!1),[b,C]=f.useState(!1),[x,y]=f.useState(""),[w,A]=f.useState(""),[E,_]=f.useState(!1),[N,D]=f.useState(c.title),[H,G]=f.useState(c.body),Q=N.trim()!==""&&H.trim()!=="",Y=async()=>{C(!0),y("");try{const U=await aw(u,N.trim(),H);A(U||"Filed."),v(!1)}catch(U){y(U.message)}finally{C(!1)}},S=()=>navigator.clipboard?.writeText(N.trim()+`

`+H).then(()=>_(!0)).catch(()=>y("Copy failed: select the text instead."));return h?a.jsxs("div",{className:"upstream-draft",children:[a.jsx("p",{className:"caption",children:"This is everything that would leave aiport: kinds and anonymous counts only, no configuration, keys, paths or Application names. A kind named by an Application is its own words: check it, and edit the text as you like. Nothing is sent until you choose below."}),a.jsxs("label",{className:"draft-field",children:["Title",a.jsx("input",{className:"draft-title",value:N,maxLength:256,disabled:g||!!w,onChange:U=>{D(U.target.value.replace(/[\r\n]/g," ")),_(!1)}})]}),a.jsxs("label",{className:"draft-field",children:["Body",a.jsx("textarea",{className:"draft-body",value:H,rows:10,disabled:g||!!w,onChange:U=>{G(U.target.value),_(!1)}})]}),x&&a.jsx("div",{className:"notice error",role:"alert",children:x}),w&&a.jsxs("div",{className:"notice",role:"status",children:["Filed with gh",w.startsWith("https://")?a.jsxs(a.Fragment,{children:[": ",a.jsx("a",{href:w,target:"_blank",rel:"noopener noreferrer",children:w})]}):"."]}),g&&!w&&a.jsxs("div",{className:"notice",role:"alertdialog","aria-label":"Confirm filing with gh",children:[a.jsxs("div",{children:["aiport will run ",a.jsx("code",{children:"gh issue create"})," on ",r.repository," with your own gh login, sending exactly the text above as you edited it. File it now?"]}),a.jsxs("div",{className:"actions",children:[a.jsx("button",{className:"primary",disabled:b,onClick:Y,children:b?"Filing…":"Yes, file it"}),a.jsx("button",{className:"secondary",disabled:b,onClick:()=>v(!1),children:"Cancel"})]})]}),!g&&!w&&a.jsxs("div",{className:"actions",children:[c.url&&Q&&a.jsxs("a",{className:"button primary with-icon",href:sw(r.repository,N.trim(),H),target:"_blank",rel:"noopener noreferrer",children:[a.jsx(dn,{size:14,"aria-hidden":!0})," Open on GitHub"]}),r.gh&&a.jsxs("button",{className:"secondary with-icon",disabled:!Q,onClick:()=>v(!0),children:[a.jsx(Vg,{size:14,"aria-hidden":!0})," File with gh…"]}),a.jsxs("button",{className:"secondary with-icon",disabled:!Q,onClick:S,children:[a.jsx(Fa,{size:14,"aria-hidden":!0})," ",E?"Copied":"Copy text"]}),a.jsx("button",{className:"text-button",onClick:()=>{m(!1),y("")},children:"Close"})]}),!c.url&&a.jsxs("p",{className:"caption",children:["No upstream repository is set: set ",a.jsx("code",{children:"upstream.repository"})," in aiport.yaml to open drafts on GitHub, or copy the text."]})]}):a.jsx("button",{className:"text-button",onClick:()=>m(!0),children:o})}const uv="aiport.tasks.chime";function Em(){try{return localStorage.getItem(uv)!=="off"}catch{return!0}}let Ct=null;function Gd(){try{Ct??=new AudioContext,Ct.state==="suspended"&&Ct.resume()}catch{}}function Om(){if(Gd(),!Ct||Ct.state!=="running")return;const o=Ct.currentTime;[[659.25,0],[987.77,.14]].forEach(([c,u])=>{const r=Ct.createOscillator(),h=Ct.createGain();r.type="sine",r.frequency.value=c,h.gain.setValueAtTime(1e-4,o+u),h.gain.exponentialRampToValueAtTime(.18,o+u+.02),h.gain.exponentialRampToValueAtTime(1e-4,o+u+.9),r.connect(h).connect(Ct.destination),r.start(o+u),r.stop(o+u+1)})}function hv(){if(Gd(),!Ct||Ct.state!=="running")return;const o=Ct.currentTime,c=Ct.createOscillator(),u=Ct.createGain();c.type="sine",c.frequency.value=523.25,u.gain.setValueAtTime(1e-4,o),u.gain.exponentialRampToValueAtTime(.08,o+.03),u.gain.exponentialRampToValueAtTime(1e-4,o+.45),c.connect(u).connect(Ct.destination),c.start(o),c.stop(o+.5)}function iw(o,c){const[u,r]=f.useState(Em),[h,m]=f.useState(null),g=f.useRef(null);return f.useEffect(()=>{const b=()=>Gd();return window.addEventListener("pointerdown",b,{once:!0}),window.addEventListener("keydown",b,{once:!0}),()=>{window.removeEventListener("pointerdown",b),window.removeEventListener("keydown",b)}},[]),f.useEffect(()=>{if(!o)return;const b=new AbortController;return fetch("/api/v1/tasks",{credentials:"same-origin",signal:b.signal}).then(C=>C.ok?C.json():null).then(C=>{if(!C)return;const x=C.tasks.filter(y=>y.status==="pending"||y.status==="in_progress").length;g.current!==null&&g.current>0&&x===0&&(m({blocked:C.tasks.filter(y=>y.status==="blocked").length}),Em()&&Om()),x>0&&m(null),g.current=x}).catch(()=>{}),()=>b.abort()},[o,c]),{enabled:u,setEnabled:b=>{r(b);try{localStorage.setItem(uv,b?"on":"off")}catch{}b&&Om()},done:h,dismiss:()=>m(null)}}function ow({enabled:o,onChange:c}){return a.jsx("button",{type:"button",className:"icon-button"+(o?" on":""),"aria-pressed":o,onClick:()=>c(!o),"aria-label":o?"Chime on":"Chime off",title:o?"Chime on: a chime plays when the queue is done. Click to mute.":"Chime off. Click to play a chime when the queue is done.",children:o?a.jsx(gx,{size:16,strokeWidth:1.75,"aria-hidden":!0}):a.jsx(mx,{size:16,strokeWidth:1.75,"aria-hidden":!0})})}function lw({blocked:o,onOpen:c,onClose:u}){return a.jsxs("div",{className:"toast",role:"status",children:[a.jsx(yx,{size:18,"aria-hidden":!0}),a.jsxs("div",{children:[a.jsx("strong",{children:"All tasks done"}),a.jsx("span",{children:o?o+" blocked task"+(o===1?"":"s")+" still need you.":"The queue is empty."})]}),a.jsx("button",{type:"button",className:"text-button",onClick:c,children:"Open tasks"}),a.jsx("button",{type:"button",className:"icon-button","aria-label":"Dismiss",onClick:u,children:a.jsx(wa,{size:14,"aria-hidden":!0})})]})}const pv="aiport.changes.sound";function fv(){try{return localStorage.getItem(pv)==="on"}catch{return!1}}function rw(o){const c=f.useRef(null),u=o?o.slice().sort().join(","):null;f.useEffect(()=>{if(!o)return;const r=c.current;c.current=new Set(o),r&&o.some(h=>!r.has(h))&&fv()&&hv()},[u])}function cw(){const[o,c]=f.useState(fv),u=()=>{const r=!o;c(r);try{localStorage.setItem(pv,r?"on":"off")}catch{}r&&hv()};return a.jsxs("button",{type:"button",className:"secondary with-icon","aria-pressed":o,onClick:u,title:o?"A short soft sound plays when something new waits for your decision. Click to turn it off.":"No sound. Click to hear a short soft sound when something new waits for your decision.",children:[o?a.jsx(Zx,{size:14,"aria-hidden":!0}):a.jsx(Kx,{size:14,"aria-hidden":!0}),o?"Sound on":"Sound off"]})}async function Fs(o,c,u="POST"){const r=await fetch("/api/v1/"+o,{method:u,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:c?JSON.stringify(c):void 0}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Request failed ("+r.status+").");return h}const dw=o=>Math.max(0,Math.ceil((o-Date.now())/6e4));function uw({p:o,targets:c,onDone:u}){const[r,h]=f.useState(""),[m,g]=f.useState(!1),[v,b]=f.useState(!1),[C,x]=f.useState(""),[y,w]=f.useState(o.repository||"");async function A(E,_){b(!0),x("");try{await Fs("pairings/"+encodeURIComponent(o.code)+"/"+E,E==="approve"?_?{replace:_}:{target:r,may_request_changes:m,repository:y.trim()}:void 0),u()}catch(N){x(N.message)}finally{b(!1)}}return a.jsxs("li",{className:"pairing",children:[a.jsxs("div",{children:[a.jsx("strong",{children:o.name})," ",a.jsx("span",{className:"pair-code","aria-label":"Code "+o.code,children:o.code}),o.description&&a.jsx("p",{className:"muted-small",children:o.description}),a.jsxs("small",{children:["Check the application shows this same code. Expires in ",dw(o.expires_at_ms)," min."]}),o.existing&&a.jsxs("p",{className:"pairing-existing",role:"note",children:[a.jsx("span",{className:"badge warn",children:"Already connected"})," as ",a.jsx("strong",{children:o.existing.name})," ",a.jsx("code",{children:o.existing.id})," (same ",o.existing.by,"). Give it a new token rather than adding a second ",o.existing.name,"."]}),o.repository&&a.jsxs("div",{className:"pairing-repository",children:[a.jsxs("label",{htmlFor:"repo-"+o.code,children:["Repository ",a.jsx("small",{className:"muted",children:"(reported by the application; confirm, edit or clear it)"})]}),a.jsx("input",{id:"repo-"+o.code,value:y,spellCheck:!1,disabled:v,onChange:E=>w(E.target.value)}),o.repository_problem&&y.trim()===o.repository&&a.jsx("small",{className:"auth-error",children:o.repository_problem})]})]}),a.jsxs("div",{className:"pairing-actions",children:[a.jsx("label",{className:"sr-only",htmlFor:"bind-"+o.code,children:"Connection"}),a.jsxs("select",{id:"bind-"+o.code,value:r,onChange:E=>h(E.target.value),title:"Optional: always send this application's calls to one connection",children:[a.jsx("option",{value:"",children:"Protocol default"}),c.filter(E=>!E.disabled).map(E=>a.jsx("option",{value:E.id,children:E.id},E.id))]}),a.jsxs("label",{className:"check-label",title:"Lets this application ask you for changes to aiport. You decide on each request.",children:[a.jsx("input",{type:"checkbox",checked:m,onChange:E=>g(E.target.checked)}),"May request changes"]}),o.existing?a.jsxs(a.Fragment,{children:[a.jsxs("button",{type:"button",className:"primary",disabled:v,onClick:()=>A("approve",o.existing.id),title:"Approve as "+o.existing.id+": it gets a new token, its old one stops working",children:["Give ",o.existing.id," a new token"]}),a.jsx("button",{type:"button",className:"secondary",disabled:v,onClick:()=>A("approve"),title:"Add another Application with its own token",children:"Add as separate"})]}):a.jsx("button",{type:"button",className:"primary",disabled:v,onClick:()=>A("approve"),children:"Approve"}),a.jsx("button",{type:"button",className:"text-button",disabled:v,onClick:()=>A("reject"),children:"Reject"})]}),C&&a.jsx("p",{className:"auth-error",role:"alert",children:C})]})}function Bd({o,onDone:c}){const[u,r]=f.useState(!1),[h,m]=f.useState("");async function g(v){r(!0),m("");try{await Fs("origins/"+encodeURIComponent(o.id),{may_request_changes:v},"PATCH"),c()}catch(b){m(b.message)}finally{r(!1)}}return a.jsxs(a.Fragment,{children:[a.jsxs("label",{className:"check-label",children:[a.jsx("input",{type:"checkbox",checked:!!o.may_request_changes,disabled:u,onChange:v=>g(v.target.checked),"aria-label":"May "+o.name+" request changes"}),o.may_request_changes?"Allowed":"Not allowed"]}),h&&a.jsx("p",{className:"auth-error",role:"alert",children:h})]})}function mv(){const[o,c]=f.useState([]),[u,r]=f.useState(0);return f.useEffect(()=>{const h=new AbortController;fetch("/api/v1/pairings",{credentials:"same-origin",signal:h.signal}).then(g=>g.ok?g.json():{pairings:[]}).then(g=>c(g.pairings||[])).catch(()=>{});const m=setTimeout(()=>r(g=>g+1),4e3);return()=>{h.abort(),clearTimeout(m)}},[u]),[o,()=>r(h=>h+1)]}function hw(o){const[c,u]=f.useState([]),[r,h]=f.useState(0);return f.useEffect(()=>{const m=new AbortController;return fetch("/api/v1/pairings?status=rejected",{credentials:"same-origin",signal:m.signal}).then(g=>g.ok?g.json():{pairings:[]}).then(g=>u(g.pairings||[])).catch(()=>{}),()=>m.abort()},[o,r]),[c,()=>h(m=>m+1)]}const pw=(o,c)=>Fs("origins",{name:o,target:c}),fw=o=>Fs("origins/"+encodeURIComponent(o)+"/token"),mw=o=>Fs("origins/"+encodeURIComponent(o),void 0,"DELETE");function gw(o){const[c,u]=f.useState([]);return f.useEffect(()=>{const r=new AbortController;return fetch("/api/v1/config",{credentials:"same-origin",signal:r.signal}).then(h=>h.ok?h.json():{}).then(h=>{u(h.missing_tokens||[]),h.app_url&&(gv=h.app_url)}).catch(()=>{}),()=>r.abort()},[o]),c}let gv="";const vw=()=>gv||location.origin;async function yw(o){try{return await navigator.clipboard.writeText(o),!0}catch{}const c=document.createElement("textarea");c.value=o,c.setAttribute("readonly",""),c.style.position="fixed",c.style.opacity="0",document.body.appendChild(c),c.select();let u=!1;try{u=document.execCommand("copy")}catch{u=!1}return c.remove(),u}function zm({text:o,label:c,primary:u}){const[r,h]=f.useState(""),m=async()=>{h(await yw(o)?"copied":"failed"),setTimeout(()=>h(""),1800)};return a.jsx("button",{type:"button",className:(u?"primary":"secondary")+" with-icon copy-action",onClick:m,"aria-label":r==="copied"?"Copied":c,children:r==="copied"?a.jsxs(a.Fragment,{children:[a.jsx(Zs,{size:14,"aria-hidden":!0})," Copied"]}):r==="failed"?a.jsx(a.Fragment,{children:"Select and copy"}):a.jsxs(a.Fragment,{children:[a.jsx(Fa,{size:14,"aria-hidden":!0})," ",c]})})}function Rm({name:o,issued:c,replaced:u,repository:r,onClose:h}){const m=r?r.replace(/\/+$/,"")+"/.aiport":".aiport",g="AIPORT_BASE_URL="+vw()+`
AIPORT_APP_TOKEN=`+c.token+`
`;return a.jsxs("section",{className:"token-card",role:"region","aria-label":"New token for "+o,children:[a.jsxs("header",{className:"token-card-head",children:[a.jsx("span",{className:"token-card-icon","aria-hidden":!0,children:a.jsx(zl,{size:16})}),a.jsxs("div",{children:[a.jsxs("h3",{children:["New token for ",o," ",a.jsx("code",{children:c.id})]}),a.jsxs("p",{children:["Copy it now: aiport will not show it again.",u?" The previous token has stopped working.":""]})]}),a.jsx("button",{type:"button",className:"icon-button token-card-close",onClick:h,"aria-label":"Close and hide the token",children:a.jsx(wa,{size:16,"aria-hidden":!0})})]}),a.jsxs("div",{className:"token-card-field",children:[a.jsx("label",{htmlFor:"token-"+c.id,children:"Token"}),a.jsxs("div",{className:"token-card-row",children:[a.jsx("input",{id:"token-"+c.id,className:"token-value",readOnly:!0,value:c.token,spellCheck:!1,onFocus:v=>v.currentTarget.select()}),a.jsx(zm,{text:c.token,label:"Copy"})]})]}),a.jsxs("div",{className:"token-card-field",children:[a.jsxs("div",{className:"token-card-label",children:[a.jsx("span",{children:".aiport"}),a.jsx(zm,{text:g,label:"Copy .aiport lines",primary:!0})]}),a.jsx("pre",{className:"token-env",children:a.jsx("code",{children:g})}),a.jsxs("small",{children:["Save as ",a.jsx("code",{children:m}),", replacing what is there; ",a.jsx("code",{children:".aiport"})," is in its ",a.jsx("code",{children:".gitignore"}),", never committed. The application and its coding agent both read it."]})]}),a.jsx("footer",{className:"token-card-foot",children:a.jsx("button",{type:"button",className:"secondary",onClick:h,children:"Done"})})]})}function xw({targets:o,onAdded:c,onCancel:u}){const[r,h]=f.useState(""),[m,g]=f.useState(""),[v,b]=f.useState(!1),[C,x]=f.useState("");async function y(w){w.preventDefault(),b(!0),x("");try{const A=await pw(r.trim(),m);c({id:A.client_id,token:A.token},r.trim())}catch(A){x(A.message)}finally{b(!1)}}return a.jsxs("form",{className:"pairing add-application",onSubmit:y,"aria-label":"Add an application with a token",children:[a.jsxs("div",{children:[a.jsxs("label",{htmlFor:"add-app-name",children:[a.jsx("strong",{children:"Add an application with a token"})," ",a.jsx("small",{className:"muted",children:"(no pairing: you paste the token into its .aiport file)"})]}),a.jsx("input",{id:"add-app-name",value:r,maxLength:128,placeholder:"Application name",disabled:v,onChange:w=>h(w.target.value),autoFocus:!0})]}),a.jsxs("div",{className:"pairing-actions",children:[a.jsx("label",{className:"sr-only",htmlFor:"add-app-target",children:"Connection"}),a.jsxs("select",{id:"add-app-target",value:m,onChange:w=>g(w.target.value),title:"Optional: always send this application's calls to one connection",children:[a.jsx("option",{value:"",children:"Protocol default"}),o.filter(w=>!w.disabled).map(w=>a.jsx("option",{value:w.id,children:w.id},w.id))]}),a.jsx("button",{type:"submit",className:"primary",disabled:v||!r.trim(),children:"Add and issue token"}),a.jsx("button",{type:"button",className:"text-button",disabled:v,onClick:u,children:"Cancel"})]}),C&&a.jsx("p",{className:"auth-error",role:"alert",children:C})]})}function ww({o,onRemoved:c,onCancel:u}){const[r,h]=f.useState(!1),[m,g]=f.useState(""),v=[];o.aliases?.length&&v.push("its own Aliases ("+o.aliases.map(C=>C.alias).join(", ")+")"),o.target&&v.push("its binding to "+o.target),(o.agent||o.repository)&&v.push("its coding agent settings"+(o.repository?" ("+o.repository+")":"")),o.may_request_changes&&v.push("its permission to request changes");async function b(){h(!0),g("");try{await mw(o.id),c()}catch(C){g(C.message),h(!1)}}return a.jsxs("div",{className:"token-reveal remove-application",role:"region","aria-label":"Remove "+o.name,children:[a.jsxs("p",{children:[a.jsxs("strong",{children:["Remove ",o.name,"?"]})," Its token stops working at once",v.length?", and "+v.join(", ")+" go with it":"",". Its activity history is kept. Reverting the removal in Changes brings it back without a token; issue a new one then."]}),a.jsxs("div",{className:"pairing-actions",children:[a.jsxs("button",{type:"button",className:"danger",disabled:r,onClick:b,children:["Remove ",o.name]}),a.jsx("button",{type:"button",className:"text-button",disabled:r,onClick:u,children:"Cancel"})]}),m&&a.jsx("p",{className:"auth-error",role:"alert",children:m})]})}const bw=o=>Fs("pairings/rejected/"+encodeURIComponent(o),void 0,"DELETE");function vv({pairings:o,targets:c,onDone:u}){return o.length?a.jsx("ul",{className:"pairings",children:o.map(r=>a.jsx(uw,{p:r,targets:c,onDone:u},r.code))}):null}function jw({origins:o,targets:c,onChange:u}){const[r,h]=mv(),m=()=>{h(),u()};return a.jsxs("section",{className:"panel grad-bottom","aria-labelledby":"apps-heading",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{id:"apps-heading",children:[a.jsx(Pm,{size:16,"aria-hidden":!0})," Applications",a.jsx(wt,{section:"Applications and pairing"})]}),a.jsx("span",{children:r.length?r.length+" waiting for approval":o.length+" connected"})]}),a.jsx(vv,{pairings:r,targets:c,onDone:m}),o.length>0?a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Application"}),a.jsx("th",{children:"Client id"}),a.jsx("th",{children:"Connection"}),a.jsx("th",{children:"Own aliases"}),a.jsx("th",{children:"May request changes"})]})}),a.jsx("tbody",{children:o.map(g=>a.jsxs("tr",{children:[a.jsx("td",{children:g.name}),a.jsx("td",{children:a.jsx("code",{children:g.id})}),a.jsx("td",{children:g.target||"Protocol default"}),a.jsx("td",{children:g.aliases?.length?g.aliases.map(v=>v.alias+" → "+v.model).join(", "):"—"}),a.jsx("td",{children:a.jsx(Bd,{o:g,onDone:m})})]},g.id))})]})}):!r.length&&a.jsx("p",{className:"quiet-empty",children:"No applications yet. Hand the Application brief on the Integrate page to the coding agent of an application; when it pairs, its request appears here and on the Integrate page for you to approve."})]})}async function Yd(o,c,u){const r=await fetch("/api/v1/"+o,{method:u,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:JSON.stringify(c)}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Request failed ("+r.status+").");return h}function yv(o){const[c,u]=f.useState({agents:[]}),[r,h]=f.useState(0);return f.useEffect(()=>{const m=new AbortController;return fetch("/api/v1/agents",{credentials:"same-origin",signal:m.signal}).then(g=>g.ok?g.json():{agents:[]}).then(u).catch(()=>{}),()=>m.abort()},[o,r]),[c,()=>h(m=>m+1)]}const xv=(o,c)=>o?.agent||c.default_agent||"",Ed=(o,c)=>c.agents.find(u=>u.id===o)?.name||o,wv=(o,c)=>!!(o?.repository&&xv(o,c));function bv({agent:o}){return!o||o.ready?null:a.jsxs("p",{className:"install-hint",role:"note",children:[o.installed?a.jsxs(a.Fragment,{children:[o.name," is installed, but aiport talks to it through its ACP adapter, which is not. Install it yourself with:"]}):a.jsxs(a.Fragment,{children:[o.name," is not installed on this machine. Install it with:"]}),a.jsx("code",{className:"install-command",children:o.install}),a.jsx("small",{className:"muted",children:"aiport never downloads or installs anything. Refresh this page afterwards."})]})}function jv({id:o,value:c,agents:u,emptyLabel:r,onChange:h,disabled:m}){return a.jsxs("select",{id:o,value:c,disabled:m,onChange:g=>h(g.target.value),children:[a.jsx("option",{value:"",children:r}),u.map(g=>a.jsxs("option",{value:g.id,disabled:!g.installed&&g.id!==c,children:[g.name,g.installed?g.ready?"":" (adapter missing)":" (not installed)"]},g.id))]})}function Aw({o,agents:c,onDone:u}){const[r,h]=f.useState(o.agent||""),[m,g]=f.useState(o.repository||""),[v,b]=f.useState(!1),[C,x]=f.useState(""),[y,w]=f.useState(!1);f.useEffect(()=>{h(o.agent||""),g(o.repository||"")},[o.id,o.agent,o.repository]),f.useEffect(()=>{x(""),w(!1)},[o.id]);const A=c.default_agent?"Default ("+Ed(c.default_agent,c)+")":"None (no default set)",E=c.agents.find(D=>D.id===(r||c.default_agent)),_=r!==(o.agent||"")||m.trim()!==(o.repository||"");async function N(){b(!0),x(""),w(!1);try{await Yd("origins/"+encodeURIComponent(o.id),{agent:r,repository:m.trim()},"PATCH"),w(!0),u()}catch(D){x(D.message)}finally{b(!1)}}return a.jsxs("div",{className:"agent-settings",children:[a.jsxs("div",{className:"agent-fields",children:[a.jsx("label",{htmlFor:"agent-"+o.id,children:"Coding agent"}),a.jsx(jv,{id:"agent-"+o.id,value:r,agents:c.agents,emptyLabel:A,onChange:h,disabled:v}),a.jsx("label",{htmlFor:"repo-"+o.id,children:"Repository"}),a.jsx("input",{id:"repo-"+o.id,value:m,disabled:v,spellCheck:!1,placeholder:"/home/you/src/docs-bot",onChange:D=>g(D.target.value),"aria-describedby":"repo-help-"+o.id})]}),a.jsxs("small",{id:"repo-help-"+o.id,className:"muted",children:["The top directory of the application's git repository (where ",a.jsx("code",{children:".git"})," is). Agent runs work only there."]}),a.jsx(bv,{agent:E}),a.jsxs("div",{className:"button-row",children:[a.jsx("button",{type:"button",className:"secondary",disabled:v||!_,onClick:N,children:v?"Saving…":"Save"}),y&&!_&&a.jsx("span",{className:"muted-small",role:"status",children:"Saved as a Change you can revert."})]}),C&&a.jsx("p",{className:"auth-error",role:"alert",children:C})]})}function kw({agents:o,targets:c,onSaved:u}){const r=o.agent_runs||{},[h,m]=f.useState(r.connection||""),[g,v]=f.useState(r.model||""),[b,C]=f.useState(!1),[x,y]=f.useState(""),[w,A]=f.useState("");f.useEffect(()=>{m(r.connection||""),v(r.model||"")},[r.connection,r.model]);const E=c.filter(D=>D.protocol==="anthropic"&&(!D.disabled||D.id===r.connection)),_=h!==(r.connection||"")||g.trim()!==(r.model||"");async function N(){C(!0),y(""),A("");try{await Yd("agents/calls",{connection:h,model:g.trim()},"PUT"),A("Saved as a Change you can revert."),u()}catch(D){y(D.message)}finally{C(!1)}}return a.jsxs("div",{className:"settings-group",children:[a.jsx("h3",{className:"subheading",children:"Model calls of Agent runs"}),a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Connection",a.jsxs("select",{id:"run-connection",value:h,disabled:b,onChange:D=>m(D.target.value),children:[a.jsx("option",{value:"",children:"Anthropic protocol default"}),E.map(D=>a.jsxs("option",{value:D.id,children:[D.id," (",D.provider,")"]},D.id))]})]}),a.jsxs("label",{children:["Model",a.jsx("input",{id:"run-model",value:g,disabled:b,maxLength:256,spellCheck:!1,autoComplete:"off",placeholder:"the agent's own choice",onChange:D=>v(D.target.value)})]})]}),a.jsxs("small",{children:["Claude Code runs call models through aiport, with a credential that lives only as long as the run, and show in Activity under the Application ",a.jsx("code",{children:"aiport"}),". Codex and Gemini CLI runs use their own login; aiport never edits an agent's configuration files. Explain runs stop after ",r.explain_minutes||10," minutes (",a.jsx("code",{children:"agent_runs.explain_minutes"}),")."]}),!E.length&&a.jsx("small",{children:"No Anthropic-protocol Connection is configured, so Claude Code runs use their own login."}),x&&a.jsx("p",{className:"auth-error",role:"alert",children:x}),a.jsxs("div",{className:"form-actions",children:[a.jsx("button",{type:"button",className:"primary",disabled:b||!_,onClick:N,children:b?"Saving…":"Save"}),w&&!_&&a.jsx("span",{className:"muted-small",role:"status",children:w})]})]})}function Cw({tick:o,targets:c,onChange:u}){const[r,h]=yv(o),[m,g]=f.useState(!1),[v,b]=f.useState("");async function C(y){g(!0),b("");try{await Yd("agents/default",{agent:y},"PUT"),h(),u()}catch(w){b(w.message)}finally{g(!1)}}const x=r.agents.find(y=>y.id===r.default_agent);return a.jsxs("section",{className:"panel","aria-labelledby":"default-agent-heading",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{id:"default-agent-heading",children:[a.jsx(Ud,{size:16,"aria-hidden":!0})," Default coding agent",a.jsx(wt,{section:"Connecting a coding agent"})]}),a.jsx("span",{children:"For Agent runs"})]}),a.jsxs("div",{className:"conn-form settings-body",children:[a.jsxs("div",{className:"settings-group",children:[a.jsx("div",{className:"form-row",children:a.jsxs("label",{children:["Default agent",a.jsx(jv,{id:"default-agent",value:r.default_agent||"",agents:r.agents,emptyLabel:"None",onChange:C,disabled:m})]})}),a.jsx("small",{children:"The coding agent aiport starts to explain or fix an Application's Issues, for every Application that has none of its own (set on the Integrate page). aiport only starts agents it found installed, by absolute path. Saved as soon as you choose."}),a.jsx(bv,{agent:x}),v&&a.jsx("p",{className:"auth-error",role:"alert",children:v})]}),a.jsx(kw,{agents:r,targets:c,onSaved:()=>{h(),u()}})]})]})}const Qs={openai:"OpenAI",anthropic:"Anthropic"},Tw={fast:"Quick, low-cost calls",balanced:"Everyday work",best:"The hardest work"},wd="\0other";function Sw(o){const[c,u]=f.useState({models:[],state:"loading"});return f.useEffect(()=>{if(!o){u({models:[],state:"unlisted"});return}const r=new AbortController;return u({models:[],state:"loading"}),fetch("/api/v1/targets/"+encodeURIComponent(o)+"/models",{credentials:"same-origin",headers:{"X-Aiport-CSRF":"1"},signal:r.signal}).then(h=>h.ok?h.json():{}).then(h=>u(h.models?.length?{models:[...h.models].sort(),state:"ok"}:{models:[],state:"unlisted"})).catch(()=>{r.signal.aborted||u({models:[],state:"unlisted"})}),()=>r.abort()},[o]),c}function Mm({id:o,label:c,value:u,onChange:r,models:h,state:m,empty:g,required:v}){const b=h.includes(u),[C,x]=f.useState(!!u&&!b&&m==="ok");f.useEffect(()=>{m==="ok"&&u&&!h.includes(u)&&x(!0)},[m]);const y=m!=="ok"||C;return a.jsxs("label",{className:"model-picker",htmlFor:o,children:[c,m==="ok"&&a.jsxs("select",{id:y?void 0:o,value:y?wd:u,onChange:w=>{w.target.value===wd?(x(!0),r(b?"":u)):(x(!1),r(w.target.value))},children:[g!==void 0?a.jsx("option",{value:"",children:g}):!u&&a.jsx("option",{value:"",disabled:!0,children:"Choose a model"}),h.map(w=>a.jsx("option",{value:w,children:w},w)),a.jsx("option",{value:wd,children:"Other model id…"})]}),y&&a.jsx("input",{id:o,value:u,onChange:w=>r(w.target.value),required:v,maxLength:256,spellCheck:!1,autoComplete:"off",placeholder:m==="loading"?"Loading the models…":g!==void 0?g:"The model id the provider expects"})]})}const Av=["low","medium","high"];function kv(o){const[c,u]=f.useState(null),[r,h]=f.useState(0);return f.useEffect(()=>{const m=new AbortController;return fetch("/api/v1/levels",{credentials:"same-origin",signal:m.signal}).then(g=>g.ok?g.json():null).then(g=>g&&u(g)).catch(()=>{}),()=>m.abort()},[o,r]),[c,()=>h(m=>m+1)]}async function qm(o,c,u){const r=await fetch("/api/v1/"+c,{method:o,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:JSON.stringify(u)}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"The change was not saved ("+r.status+").");return h}const _l=(o,c,u)=>o?.find(r=>r.level===c&&r.protocol===u);function Dm({m:o,doc:c}){const u=c.connections.find(h=>h.id===o.connection),r=Av.filter(h=>o.effort?.[h]);return a.jsxs("span",{className:"level-mapping",children:[a.jsxs("span",{className:"level-model",children:[u&&a.jsx(Ks,{provider:u.provider,location:u.location,size:16}),a.jsx("code",{children:o.model})]}),a.jsxs("small",{className:"muted",children:["on ",o.connection]}),r.length>0&&a.jsx("span",{className:"level-efforts-chips",children:r.map(h=>a.jsxs("span",{className:"effort-chip",title:"Calls at "+h+" effort get "+o.effort[h],children:[h,": ",a.jsx("code",{children:o.effort[h]})]},h))})]})}function Nw({doc:o,level:c,protocol:u,application:r,current:h,defaults:m,onDone:g,onCancel:v}){const b=o.connections.filter($=>$.protocol===u&&(!$.disabled||$.id===h?.connection)),C=u==="anthropic"?m?.anthropic_target:m?.openai_target,[x,y]=f.useState(h?.connection||b.find($=>$.id===C)?.id||b[0]?.id||""),w=Sw(x),[A,E]=f.useState(h?.model||""),[_,N]=f.useState(h?.effort||{}),[D,H]=f.useState(!1),[G,Q]=f.useState(""),Y=b.find($=>$.id===x),S=c+" on "+Qs[u];async function U($){$.preventDefault(),H(!0),Q("");try{await qm("PUT","levels/"+u+"/"+c,{application:r,connection:x,model:A.trim(),effort:_}),g("Saved "+S+" as a Change. Revert it in Changes.")}catch(oe){Q(oe.message)}finally{H(!1)}}async function Z(){H(!0),Q("");try{await qm("DELETE","levels/"+u+"/"+c,{application:r}),g(r?"Cleared "+r+"'s own "+S+": it uses the shared mapping. Revert it in Changes.":"Cleared the shared "+S+". Revert it in Changes.")}catch($){Q($.message)}finally{H(!1)}}return a.jsxs("form",{className:"conn-form level-editor",onSubmit:U,"aria-label":"Map "+S,children:[a.jsxs("h3",{children:[r?r+"'s own ":"Shared ",a.jsx("code",{children:c})," on ",Qs[u]]}),b.length===0?a.jsxs("p",{className:"muted",children:["No ",Qs[u]," Connection yet. Add one on Providers first."]}):a.jsxs(a.Fragment,{children:[a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Connection",a.jsx("select",{value:x,onChange:$=>{y($.target.value),E(""),N({})},children:b.map($=>a.jsxs("option",{value:$.id,children:[$.id," · ",$.provider,$.id===C?" (protocol default)":""]},$.id))})]}),a.jsx(Mm,{id:"level-model-"+c+u,label:"Model",value:A,onChange:E,models:w.models,state:w.state,required:!0},"model-"+x)]}),w.state==="unlisted"&&a.jsxs("small",{className:"muted",children:[Y?.id," does not list its models; type the id the provider expects."]}),a.jsxs("fieldset",{children:[a.jsxs("legend",{children:["Per-effort models ",a.jsx("span",{className:"muted",children:"(optional)"})]}),a.jsx("div",{className:"level-efforts",children:Av.map($=>a.jsx(Mm,{id:"level-"+$+c+u,label:$[0].toUpperCase()+$.slice(1)+" effort",value:_[$]||"",onChange:oe=>N(be=>({...be,[$]:oe})),models:w.models,state:w.state,empty:"Same as the model"+(A.trim()?" ("+A.trim()+")":"")},$+x))}),a.jsxs("small",{className:"muted",children:["A call that says how hard it is (",a.jsx("code",{children:"X-Aiport-Effort"}),", or its provider's own effort field) gets the model of its effort."]})]})]}),G&&a.jsx("p",{className:"auth-error",role:"alert",children:G}),a.jsxs("div",{className:"form-actions",children:[b.length>0&&a.jsx("button",{className:"primary",disabled:D||!x||!A.trim(),children:D?"Saving…":"Save"}),h&&a.jsx("button",{type:"button",className:"secondary",disabled:D,onClick:Z,children:r?"Use the shared mapping":"Clear"}),a.jsx("button",{type:"button",className:"text-button",onClick:v,children:"Cancel"})]})]})}function Cv({doc:o,application:c,defaults:u,onChange:r}){const[h,m]=f.useState(null),[g,v]=f.useState(""),b=c?o.applications[c]||[]:o.shared,C=(x,y)=>Object.values(o.applications).filter(w=>_l(w,x,y)).length;return a.jsxs(a.Fragment,{children:[a.jsx("div",{className:"levels-table",children:a.jsxs("table",{"aria-label":c?"Levels of "+c:"Shared Levels",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Level"}),o.protocols.map(x=>a.jsx("th",{children:Qs[x]||x},x))]})}),a.jsx("tbody",{children:o.levels.map(x=>a.jsxs("tr",{children:[a.jsx("th",{scope:"row",children:a.jsxs("span",{className:"level-name",children:[a.jsx("code",{children:x}),a.jsx("small",{className:"muted",children:Tw[x]||""})]})}),o.protocols.map(y=>{const w=_l(b,x,y),A=c?_l(o.shared,x,y):void 0,E=h?.level===x&&h.protocol===y,_=c?0:C(x,y);return a.jsxs("td",{className:E?"editing":void 0,children:[a.jsxs("div",{className:"level-cell",children:[w?a.jsxs(a.Fragment,{children:[c&&a.jsx("span",{className:"badge info",children:"Own"}),a.jsx(Dm,{m:w,doc:o})]}):A?a.jsxs("span",{className:"level-shared",children:[a.jsx("small",{className:"muted",children:"Shared"}),a.jsx(Dm,{m:A,doc:o})]}):a.jsxs("span",{className:"level-unmapped",children:[a.jsx("span",{className:"badge warn",children:"Not mapped"}),a.jsxs("small",{className:"muted",children:["Calls sending ",a.jsx("code",{children:x})," are refused"]})]}),a.jsx("button",{type:"button",className:"row-action"+(!w&&!A&&!E?" level-map":""),"aria-pressed":E,onClick:()=>{v(""),m(E?null:{level:x,protocol:y})},title:E?"Close":c&&!w?"Give "+c+" its own "+x:(w?"Edit ":"Map ")+x+" on "+Qs[y],"aria-label":(E?"Close the editor of ":c&&!w?"Override ":w?"Edit ":"Map ")+x+" on "+Qs[y],children:E?a.jsxs(a.Fragment,{children:[a.jsx(wa,{size:14,"aria-hidden":!0})," Close"]}):w?a.jsxs(a.Fragment,{children:[a.jsx(Nd,{size:14,"aria-hidden":!0})," Edit"]}):c&&A?a.jsxs(a.Fragment,{children:[a.jsx(Nd,{size:14,"aria-hidden":!0})," Override"]}):a.jsxs(a.Fragment,{children:[a.jsx(ql,{size:14,"aria-hidden":!0})," Map"]})})]}),_>0&&a.jsx("small",{className:"muted",children:_===1?"1 Application has its own":_+" Applications have their own"})]},y)})]},x))})]})}),h&&a.jsx(Nw,{doc:o,level:h.level,protocol:h.protocol,application:c,defaults:u,current:_l(b,h.level,h.protocol),onCancel:()=>m(null),onDone:x=>{m(null),v(x),r()}},h.level+h.protocol),g&&a.jsx("p",{className:"auth-success",role:"status",children:g})]})}function _w({refresh:o,defaults:c,onChange:u}){const[r,h]=kv(o);return a.jsxs("section",{className:"panel levels-panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:[a.jsx(Nx,{size:16,"aria-hidden":!0})," Levels",a.jsx(wt,{section:"Levels"})]}),a.jsxs("span",{children:["What ",a.jsx("code",{children:"fast"}),", ",a.jsx("code",{children:"balanced"})," and ",a.jsx("code",{children:"best"})," mean on this aiport"]})]}),a.jsx("p",{className:"muted",children:"Applications send a Level as their model instead of naming a provider. Map each one per protocol; an Application can have its own mapping on Integrate, which comes before this one."}),r?a.jsx(Cv,{doc:r,application:"",defaults:c,onChange:()=>{h(),u()}}):a.jsx("p",{className:"quiet-empty",children:"Loading Levels…"})]})}function Ew(o,c){const u=o?.applications[c]?.length||0;return o?u===0?"Shared":u===1?"1 own":u+" own":"…"}function Ow({doc:o,application:c,onChange:u}){return a.jsxs("div",{className:"app-levels",children:[a.jsxs("p",{className:"muted",children:[c,"'s own mapping of a Level comes before the shared one, on that protocol only.",a.jsx(wt,{section:"Levels"})]}),a.jsx(Cv,{doc:o,application:c,onChange:u})]})}const Ys="",Rl=6,zw=[["python","Python"],["typescript","TypeScript"],["curl","curl"]],Rw=[["approved","Approved"],["rejected","Rejected"],["all","All"]],bd="aiport.integrate.language",Tv="aiport.integrate.mcp-copied.";function Fi(o,c){try{return c!==void 0&&localStorage.setItem(o,c),localStorage.getItem(o)||""}catch{return c||""}}const Ml=o=>o.replace(/^https?:\/\/[^/]+/,"");function Ka({text:o,label:c,onCopied:u}){const[r,h]=f.useState(!1),m=()=>{navigator.clipboard?.writeText(o).then(()=>{h(!0),u?.(),setTimeout(()=>h(!1),1500)})};return a.jsx("button",{type:"button",className:"secondary with-icon copy-button",disabled:!o,onClick:m,"aria-label":r?"Copied":c,children:r?a.jsxs(a.Fragment,{children:[a.jsx(Zs,{size:14,"aria-hidden":!0})," Copied"]}):a.jsxs(a.Fragment,{children:[a.jsx(Fa,{size:14,"aria-hidden":!0})," ",c]})})}function Mw({s:o,onCopied:c}){return a.jsxs("figure",{className:"snippet",children:[a.jsxs("figcaption",{children:[a.jsx("span",{children:o.title}),o.protocol&&!o.available&&a.jsxs("span",{className:"badge warn",title:"Ask the Owner to configure this protocol first",children:[o.protocol," not configured"]}),a.jsx(Ka,{text:o.code,label:"Copy",onCopied:c})]}),a.jsx("pre",{className:"snippet-code",children:a.jsx("code",{children:o.code})})]})}function qw(o,c){const[u,r]=f.useState({off:!1,at:{}}),h=o.join(`
`);return f.useEffect(()=>{if(!o.length){r({off:!1,at:{}});return}const m=new AbortController,g={credentials:"same-origin",signal:m.signal};return Promise.all([fetch("/api/v1/stats?hours=1",g).then(v=>v.ok?v.json():{}),...o.map(v=>fetch("/api/v1/requests?"+new URLSearchParams({origin:v,hours:"168",limit:"1"}),g).then(b=>b.ok?b.json():{requests:[]}).then(b=>[v,b.requests?.[0]?.started_at_ms??null]))]).then(([v,...b])=>r({off:v.telemetry?.state==="disabled",at:Object.fromEntries(b)})).catch(()=>{}),()=>m.abort()},[h,c]),u}function Dw(o){const[c,u]=f.useState({}),[r,h]=f.useState(null),[m,g]=f.useState(0);return f.useEffect(()=>{const v=new AbortController;return fetch("/api/v1/issues",{credentials:"same-origin",signal:v.signal}).then(b=>b.ok?b.json():{issues:[]}).then(b=>{const C={};for(const x of b.issues||[])(C[x.application]||=[]).push(x);u(C),h(zv(b))}).catch(()=>{}),()=>v.abort()},[o,m]),{issues:c,explain:r,reload:()=>g(v=>v+1)}}function Lw(o){const c=Math.min(o.never_confirmed||0,o.behind),u=o.behind-c,r=m=>m===1?"1 Application's coding agent":m+" Applications' coding agents",h=[];return u&&h.push(r(u)+(u===1?" is":" are")+" behind: "+(u===1?"it confirmed":"they confirmed")+" an older guidance version than "+o.current),c&&h.push(r(c)+(c===1?" has":" have")+" never confirmed any guidance version"),h.join("; ")+"."}function Sv(o){const[c,u]=f.useState(null);return f.useEffect(()=>{const r=new AbortController;return fetch("/api/v1/guidance",{credentials:"same-origin",signal:r.signal}).then(h=>h.ok?h.json():null).then(u).catch(()=>{}),()=>r.abort()},[o]),c}function Uw({text:o,label:c}){const[u,r]=f.useState(!1),h=()=>{navigator.clipboard?.writeText(o).then(()=>{r(!0),setTimeout(()=>r(!1),1500)})};return a.jsx("button",{type:"button",className:"row-action icon",onClick:h,"aria-label":u?"Copied":c,title:c,children:u?a.jsx(Zs,{size:14,"aria-hidden":!0}):a.jsx(Fa,{size:14,"aria-hidden":!0})})}function Lm({a:o}){if(!o.calls)return null;const c=(u,r,h,m)=>a.jsxs("span",{className:"std-chip "+(u?"good":"warn"),title:u?h:m,children:[u?a.jsx(Zs,{size:11,"aria-hidden":!0}):"!"," ",r]});return a.jsxs("span",{className:"std-chips",children:[c(o.sends_modules,"Modules","Its calls name their module (X-Aiport-Module).","None of its calls in the last 7 days named a module (X-Aiport-Module): the Feedback page cannot tell its parts apart. Hand its coding agent the communication standard."),c(o.sends_votes,"Votes","Its users' thumbs up and down arrive.","No votes in the last 7 days: add the thumbs up and down the communication standard asks for, or nobody has voted yet.")]})}function Hw({id:o,guidance:c}){const u=c?.applications[o];if(!c||!u)return a.jsx("span",{className:"muted",children:"…"});if(!u.behind)return a.jsxs("div",{className:"guidance-cell",children:[a.jsx("span",{className:"badge good",title:u.confirmed_at_ms?"Confirmed "+new Date(u.confirmed_at_ms).toLocaleString():void 0,children:"Up to date"}),a.jsx(Lm,{a:u})]});const r=u.confirmed?"Behind v"+u.confirmed+" → v"+c.current:"Unknown",h=u.confirmed?"Its coding agent confirmed guidance version "+u.confirmed+"; aiport's is "+c.current+".":"Its coding agent has never confirmed a guidance version; aiport's is "+c.current+".";return a.jsxs("div",{className:"guidance-cell",children:[a.jsx("span",{className:"badge"+(u.confirmed?" warn":" info"),title:h+(u.sends_levels?" Its calls sent Levels in the last 6 hours, a hint that its code may follow version 2.":""),children:r}),u.sends_levels&&a.jsx("small",{className:"muted",children:"Sends Levels"}),a.jsx(Uw,{text:c.instruction,label:"Copy the refresh instruction for its coding agent"}),a.jsx(Lm,{a:u})]})}const Od=o=>o.cause==="unknown_model"?"Unknown model "+o.subject:o.cause==="failing_connection"?"Failing Connection "+o.subject:o.cause==="level_unmapped"?"Unmapped Level "+o.subject:Pi(o)?"Catalog change "+o.subject:o.cause==="reported_request"?"Reported request "+(o.report?.request_id||""):o.summary;function Iw({issues:o,explain:c,onOpen:u,onExplained:r}){return a.jsx("ul",{className:"app-issues",children:o.map(h=>a.jsxs("li",{children:[a.jsxs("div",{className:"app-issue-head",children:[a.jsx("span",{className:"badge"+(h.state==="resolving"?" good":" warn"),children:h.state==="resolving"?"Resolving":Pi(h)?"Catalog change":"Issue"}),a.jsx("span",{children:h.summary}),a.jsxs("button",{type:"button",className:"text-button",onClick:()=>u(h),"aria-label":"Open the Issue "+Od(h)+" in Changes",children:["Open in Changes ",a.jsx(dn,{size:14,"aria-hidden":!0})]})]}),Pi(h)?a.jsx(Ov,{issue:h}):a.jsx(Mv,{issue:h,explain:c,onExplained:r}),a.jsx(qv,{issue:h,explain:c,onChanged:r}),a.jsx(Lv,{issue:h})]},h.id))})}function Nv(o,c){return[!!o||c.handed||c.waiting,!!o,typeof c.seen=="number",!!o?.may_request_changes&&c.adapting,c.mcp,wv(o,c.agents)]}const zd=o=>Fi(Tv+o)==="1";function Is({n:o,done:c,current:u,title:r,children:h}){return a.jsxs("li",{className:"step"+(c?" done":"")+(u?" current":""),"aria-current":u?"step":void 0,children:[a.jsx("span",{className:"step-number","aria-hidden":!0,children:c?a.jsx(ug,{size:20,strokeWidth:1.75}):u?o:a.jsx(bx,{size:20,strokeWidth:1.75})}),a.jsxs("div",{children:[a.jsxs("strong",{children:[o,". ",r]}),a.jsx("span",{className:"sr-only",children:c?" (done)":u?" (do this now)":" (not yet)"}),a.jsx("div",{className:"step-body",children:h})]})]})}const _v=o=>{const c=Math.round((Date.now()-o)/6e4);return c<1?"just now":c<60?c+" min ago":c<2880?Math.round(c/60)+" h ago":Math.round(c/1440)+" days ago"};function Gw({x:o,onDone:c}){const[u,r]=f.useState(!1),[h,m]=f.useState("");async function g(){r(!0),m("");try{await bw(o.id),c()}catch(v){m(v.message)}finally{r(!1)}}return a.jsxs("tr",{className:"rejected-row",children:[a.jsxs("td",{children:[a.jsx("strong",{children:o.name}),o.description&&a.jsx("small",{className:"muted",children:o.description})]}),a.jsx("td",{children:a.jsx("span",{className:"pair-code small",children:o.code})}),a.jsx("td",{children:a.jsx("span",{className:"badge bad",children:"Rejected"})}),a.jsx("td",{title:new Date(o.rejected_at_ms).toLocaleString(),children:_v(o.rejected_at_ms)}),a.jsxs("td",{className:"row-actions",children:[a.jsx("button",{type:"button",className:"text-button",disabled:u,onClick:g,"aria-label":"Forget "+o.name,children:"Forget"}),h&&a.jsx("small",{className:"auth-error",role:"alert",children:h})]})]})}function Bw({origins:o,targets:c,providers:u,pairings:r,rejected:h,calls:m,issues:g,guidance:v,explain:b,onExplained:C,agents:x,levels:y,onLevels:w,app:A,onSelect:E,onAgent:_,onNew:N,adapting:D,onDone:H,onForgot:G,onIssue:Q,missing:Y}){const[S,U]=f.useState(""),Z=B=>{U(Re=>Re===B?"":B),E(B)},[$,oe]=f.useState(!1),[be,re]=f.useState(""),[J,ve]=f.useState(""),[W,F]=f.useState(null),[ne,le]=f.useState(""),[je,Ae]=f.useState({});async function Qe(B){ve(""),le(B.id),Ae(Re=>({...Re,[B.id]:""}));try{const Re=await fw(B.id);F({id:B.id,token:Re.token,name:B.name,replaced:!Y.includes(B.id)}),H()}catch(Re){Ae(Ie=>({...Ie,[B.id]:Re.message}))}finally{le("")}}const[Ze,dt]=f.useState("approved"),[j,I]=f.useState({}),[ee,ie]=f.useState({}),[fe,ye]=f.useState({}),xe=JSON.stringify([o,c]),ae={approved:o.length,rejected:h.length,all:o.length+h.length},he=Ze!=="rejected",ut=Ze!=="approved",Qt=B=>m.off?"Not recorded":m.at[B]===void 0?"…":m.at[B]===null?"None in 7 days":_v(m.at[B]);return a.jsxs("section",{className:"panel apps-list grad-bottom","aria-labelledby":"apps-heading",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{id:"apps-heading",children:[a.jsx(Pm,{size:16,"aria-hidden":!0})," Applications",a.jsx(wt,{section:"Applications"})]}),a.jsxs("div",{className:"apps-tools",children:[a.jsx("div",{className:"segmented apps-filter",role:"radiogroup","aria-label":"Show applications",children:Rw.map(([B,Re])=>a.jsxs("button",{type:"button",role:"radio","aria-checked":Ze===B,className:Ze===B?"on":"",onClick:()=>dt(B),children:[Re," ",a.jsx("span",{className:"count",children:ae[B]})]},B))}),a.jsxs("button",{type:"button",className:"secondary with-icon",onClick:N,"aria-pressed":A===Ys,children:[a.jsx(ql,{size:14,"aria-hidden":!0})," Onboard a new application"]}),a.jsxs("button",{type:"button",className:"secondary with-icon",onClick:()=>oe(B=>!B),"aria-expanded":$,title:"Create an application and its token yourself, without pairing, and paste the token into its .aiport file",children:[a.jsx(zl,{size:14,"aria-hidden":!0})," Add with a token"]})]})]}),r.length>0&&a.jsxs("div",{className:"waiting",role:"region","aria-label":"Pairing requests waiting",children:[a.jsxs("h3",{className:"waiting-heading",children:[a.jsx("span",{className:"dot amber","aria-hidden":!0})," ",r.length===1?"1 pairing waiting for approval":r.length+" pairings waiting for approval"]}),a.jsx(vv,{pairings:r,targets:c,onDone:H})]}),$&&a.jsx(xw,{targets:c,onCancel:()=>oe(!1),onAdded:(B,Re)=>{oe(!1),F({...B,name:Re,replaced:!1}),H()}}),W&&!o.some(B=>B.id===W.id)&&a.jsx(Rm,{name:W.name,issued:W,replaced:W.replaced,onClose:()=>F(null)}),he&&o.length>0&&a.jsx("div",{className:"apps-table",children:a.jsxs("table",{"aria-label":"Approved applications",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Application"}),a.jsx("th",{children:"Connection"}),a.jsx("th",{title:"May request changes",children:"Changes"}),a.jsx("th",{children:"Last call"}),a.jsx("th",{children:"Issues"}),a.jsx("th",{title:"The aiport guidance version its coding agent last confirmed",children:"Guidance"}),a.jsx("th",{title:"Coding agent for Agent runs",children:"Agent"}),a.jsx("th",{children:"Onboarding"}),a.jsx("th",{children:a.jsx("span",{className:"sr-only",children:"Actions"})})]})}),a.jsx("tbody",{children:o.map(B=>{const Re=Nv(B,{handed:!0,waiting:!1,seen:m.at[B.id],adapting:D,mcp:zd(B.id),agents:x}).filter(Boolean).length,Ie=A===B.id,we=g[B.id]||[],st=!!j[B.id]&&we.length>0,de=!!fe[B.id],Fe=!!ee[B.id]&&!!y;return a.jsxs($i.Fragment,{children:[a.jsxs("tr",{className:"app-row"+(Ie?" selected":""),onClick:Ve=>{Ve.target.closest("input,button,a,label")||Z(B.id)},children:[a.jsxs("td",{className:"app-cell",children:[a.jsxs("div",{className:"app-name",children:[a.jsx("button",{type:"button",className:"row-select","aria-expanded":S===B.id,onClick:()=>Z(B.id),title:"Open "+B.name+"'s token and removal actions",children:B.name}),a.jsx("code",{children:B.id}),Y.includes(B.id)&&a.jsx("span",{className:"badge warn",title:"Its calls are refused until you issue a token",children:"No token"})]}),a.jsxs("span",{className:"app-toggles",children:[a.jsxs("button",{type:"button",className:"text-button issue-count app-catalog-toggle","aria-expanded":de,"aria-controls":"app-catalog-"+B.id,onClick:()=>ye(Ve=>({...Ve,[B.id]:!Ve[B.id]})),title:"What "+B.name+" can send, what serves it, and its recent Catalog changes",children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"})," Catalog"]}),a.jsxs("button",{type:"button",className:"text-button issue-count app-catalog-toggle","aria-expanded":Fe,"aria-controls":"app-levels-"+B.id,disabled:!y,onClick:()=>ie(Ve=>({...Ve,[B.id]:!Ve[B.id]})),title:B.name+"'s own mapping of fast, balanced and best, before the shared one: "+Ew(y,B.id),children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"})," Levels"]})]}),je[B.id]&&a.jsx("small",{className:"auth-error",role:"alert",children:je[B.id]})]}),a.jsx("td",{children:B.target?a.jsx("code",{children:B.target}):a.jsx("span",{className:"muted",children:"Protocol default"})}),a.jsx("td",{children:a.jsx(Bd,{o:B,onDone:H})}),a.jsx("td",{title:typeof m.at[B.id]=="number"?new Date(m.at[B.id]).toLocaleString():void 0,children:Qt(B.id)}),a.jsx("td",{children:we.length?a.jsxs("button",{type:"button",className:"text-button issue-count","aria-expanded":st,"aria-controls":"app-issues-"+B.id,onClick:()=>I(Ve=>({...Ve,[B.id]:!Ve[B.id]})),title:we.map(Od).join(`
`),children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),a.jsx("span",{className:"badge warn",children:we.length})," ",we.length===1?Od(we[0]):we.length+" Issues"]}):a.jsx("span",{className:"muted",children:"None"})}),a.jsx("td",{children:a.jsx(Hw,{id:B.id,guidance:v})}),a.jsx("td",{children:(()=>{const Ve=xv(B,x),nn=wv(B,x);return a.jsx("button",{type:"button",className:"text-button",onClick:()=>_(B.id),"aria-label":(nn?"Change the coding agent of ":"Connect a coding agent to ")+B.name,title:nn?Ed(Ve,x)+(B.agent?"":" (default)")+" in "+B.repository:"Connect a coding agent (onboarding step 6)",children:nn?Ed(Ve,x):"Connect"})})()}),a.jsxs("td",{children:[a.jsx("span",{className:"progress"+(Re===Rl?" complete":""),"aria-hidden":!0,children:a.jsx("i",{style:{width:Re/Rl*100+"%"}})})," ",a.jsxs("span",{className:"progress-text",children:[a.jsx("span",{className:"sr-only",children:"Onboarding steps done: "}),Re,"/",Rl]})]}),a.jsxs("td",{className:"row-actions",children:[a.jsxs("button",{type:"button",className:"row-action",disabled:ne===B.id,onClick:()=>Y.includes(B.id)?Qe(B):ve(Ve=>Ve===B.id?"":B.id),"aria-expanded":Y.includes(B.id)?void 0:J===B.id,"aria-label":"Issue a new token for "+B.name,title:"Issue a new token for "+B.name+" to paste into its .aiport file; the current one stops working",children:[a.jsx(zl,{size:14,"aria-hidden":!0})," ",Y.includes(B.id)?"Issue token":"New token"]}),a.jsx("button",{type:"button",className:"row-action icon remove",onClick:()=>re(Ve=>Ve===B.id?"":B.id),"aria-expanded":be===B.id,"aria-label":"Remove "+B.name,title:"Remove "+B.name,children:a.jsx(Dl,{size:14,"aria-hidden":!0})})]})]}),S===B.id&&a.jsx("tr",{className:"app-issues-row",children:a.jsx("td",{colSpan:9,children:a.jsxs("div",{className:"app-actions",role:"region","aria-label":B.name+" actions",children:[a.jsxs("p",{children:[a.jsx("strong",{children:B.name})," ",a.jsx("code",{children:B.id}),Y.includes(B.id)?a.jsxs(a.Fragment,{children:[" · ",a.jsx("span",{className:"badge warn",children:"No token"})," its calls are refused until you issue one."]}):a.jsxs(a.Fragment,{children:[" · it reads its token from ",a.jsx("code",{children:".aiport"})," at its repository root."]})]}),a.jsxs("div",{className:"pairing-actions",children:[a.jsxs("button",{type:"button",className:"secondary with-icon",disabled:ne===B.id,onClick:()=>Y.includes(B.id)?Qe(B):ve(B.id),children:[a.jsx(zl,{size:14,"aria-hidden":!0})," ",Y.includes(B.id)?"Issue token":"Issue new token"]}),a.jsxs("button",{type:"button",className:"danger with-icon",onClick:()=>re(B.id),children:[a.jsx(Dl,{size:14,"aria-hidden":!0})," Remove application"]}),a.jsx("button",{type:"button",className:"text-button",onClick:()=>U(""),children:"Close"})]})]})})}),W?.id===B.id&&a.jsx("tr",{className:"app-issues-row",children:a.jsx("td",{colSpan:9,children:a.jsx(Rm,{name:B.name,issued:W,replaced:W.replaced,repository:B.repository,onClose:()=>F(null)})})}),J===B.id&&a.jsx("tr",{className:"app-issues-row",children:a.jsx("td",{colSpan:9,children:a.jsxs("div",{className:"token-reveal",role:"region","aria-label":"New token for "+B.name,children:[a.jsxs("p",{children:[a.jsxs("strong",{children:["Issue a new token for ",B.name,"?"]})," Its current token stops working at once; put the new one in its ",a.jsx("code",{children:".aiport"})," file."]}),a.jsxs("div",{className:"pairing-actions",children:[a.jsx("button",{type:"button",className:"primary",disabled:ne===B.id,onClick:()=>Qe(B),children:"Issue new token"}),a.jsx("button",{type:"button",className:"text-button",onClick:()=>ve(""),children:"Cancel"})]})]})})}),be===B.id&&a.jsx("tr",{className:"app-issues-row",children:a.jsx("td",{colSpan:9,children:a.jsx(ww,{o:B,onCancel:()=>re(""),onRemoved:()=>{re(""),W?.id===B.id&&F(null),H()}})})}),st&&a.jsx("tr",{className:"app-issues-row",id:"app-issues-"+B.id,children:a.jsx("td",{colSpan:9,children:a.jsx(Iw,{issues:we,explain:b,onOpen:Q,onExplained:C})})}),de&&a.jsx("tr",{className:"app-issues-row",id:"app-catalog-"+B.id,children:a.jsx("td",{colSpan:9,children:a.jsx(ob,{application:B.id,targets:c,providers:u,reload:xe})})}),Fe&&a.jsx("tr",{className:"app-issues-row",id:"app-levels-"+B.id,children:a.jsx("td",{colSpan:9,children:a.jsx(Ow,{doc:y,application:B.id,onChange:w})})})]},B.id)})})]})}),ut&&h.length>0&&a.jsx("div",{className:"apps-table rejected-table",children:a.jsxs("table",{"aria-label":"Rejected applications",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Rejected application"}),a.jsx("th",{children:"Code"}),a.jsx("th",{children:"Status"}),a.jsx("th",{children:"Rejected"}),a.jsx("th",{children:a.jsx("span",{className:"sr-only",children:"Actions"})})]})}),a.jsx("tbody",{children:h.map(B=>a.jsx(Gw,{x:B,onDone:G},B.id))})]})}),!r.length&&!(he&&o.length)&&!(ut&&h.length)&&a.jsx("p",{className:"quiet-empty",children:Ze==="rejected"?"No rejected applications. A pairing request you reject is listed here, so you can see who asked.":"No applications yet. Onboard one below: hand the instruction to the coding agent of an application; when it pairs, its request appears here for you to approve."})]})}function Yw({origins:o,app:c,setApp:u,pairings:r,seen:h,off:m,adapting:g,agents:v,prompt:b,docs:C,mcp:x,setMcp:y,open:w,setOpen:A,onDone:E,onAgentBrief:_}){const[N,D]=f.useState(!1),H=o.find(U=>U.id===c),G=Nv(H,{handed:N,waiting:r.length>0,seen:h,adapting:g,mcp:x,agents:v}),Q=G.findIndex(U=>!U),Y=G.filter(Boolean).length,S=r.length?r.length===1?"1 pairing waiting":r.length+" pairings waiting":H?Q<0?H.name+" is onboarded":H.name+" · "+Y+"/"+Rl:"A new application";return a.jsxs("section",{className:"panel wizard","aria-labelledby":"wizard-heading",children:[a.jsxs("div",{className:"panel-heading wizard-heading",children:[a.jsxs("h2",{id:"wizard-heading",children:[a.jsxs("button",{type:"button",className:"group-toggle","aria-expanded":w,"aria-controls":"wizard-body",onClick:()=>A(!w),children:[a.jsx(un,{size:16,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),"Onboard an application"]}),a.jsx(wt,{section:"Onboarding an application, step by step"})]}),a.jsx("span",{className:"badge"+(r.length?" warn":Q<0&&H?" good":" info"),children:S})]}),w&&a.jsxs("div",{id:"wizard-body",children:[a.jsxs("p",{className:"wizard-follows muted-small",children:["Following ",a.jsx("strong",{children:H?H.name:"a new application"}),H&&a.jsxs(a.Fragment,{children:[" · ",a.jsx("button",{type:"button",className:"text-button",onClick:()=>u(Ys),children:"Onboard a new application instead"})]}),". Select an application in the list to follow it."]}),a.jsxs("ol",{className:"steps",children:[a.jsxs(Is,{n:1,done:G[0],current:Q===0,title:"Hand this to the coding agent",children:[a.jsxs("p",{children:["Open the coding agent (Claude Code, Codex, Cursor…) ",a.jsx("strong",{children:"in the application's repository"})," and paste this instruction. It tells the agent to read ",a.jsx("a",{href:Ml(C),target:"_blank",rel:"noreferrer",children:a.jsx("code",{children:C})}),", which aiport writes from this gateway and which holds everything for both parts: wiring the application's LLM calls through aiport, and letting later sessions ask aiport for changes."]}),a.jsx("blockquote",{className:"agent-prompt",children:b}),a.jsxs("div",{className:"button-row",children:[a.jsx(Ka,{text:b,label:"Copy instruction",onCopied:()=>D(!0)}),a.jsx(Ka,{text:C,label:"Copy URL only"})]})]}),a.jsx(Is,{n:2,done:G[1],current:Q===1,title:"Approve the pairing",children:H?a.jsxs(a.Fragment,{children:["Connected as ",a.jsx("code",{children:H.id}),H.target?a.jsxs(a.Fragment,{children:[", bound to ",a.jsx("code",{children:H.target})]}):", using the protocol defaults","."]}):r.length>0?a.jsxs(a.Fragment,{children:["The agent asked to pair: the request waits at the top of ",a.jsx("strong",{children:"Applications"})," above, with code",r.length>1?"s":""," ",r.map(U=>a.jsx("span",{className:"pair-code small",children:U.code},U.code)),". Check it matches the code the agent shows you, then approve it there; you can bind a Connection and allow change requests."]}):a.jsx(a.Fragment,{children:"Waiting for the agent to ask. Its request appears at the top of Applications within a few seconds, with a short code; requests wait 30 minutes and survive restarts."})}),a.jsx(Is,{n:3,done:G[2],current:Q===2,title:"First call arrives",children:H?m?a.jsxs(a.Fragment,{children:["Activity recording is off (",a.jsx("code",{children:"telemetry.enabled: false"}),"), so aiport cannot tell. Ask the agent whether its test call succeeded."]}):h===void 0?"Checking…":typeof h=="number"?a.jsxs(a.Fragment,{children:["Latest call ",new Date(h).toLocaleString(),". The application now goes through aiport."]}):"No call from this application in the last 7 days yet. If its calls show as unknown in Activity, it is not sending its own token.":"After pairing, the agent makes one real call through aiport."}),a.jsx(Is,{n:4,done:G[3],current:Q===3,title:"Allow change requests (optional)",children:H?g?a.jsxs(a.Fragment,{children:[a.jsx(Bd,{o:H,onDone:E})," You still decide each request on the Changes page."]}):a.jsxs(a.Fragment,{children:["Adapting is turned off in the configuration (",a.jsx("code",{children:"adapting.disabled"}),"), so no Application can request changes now."]}):"Once paired, decide whether coding agents in its repository may ask you for changes."}),a.jsx(Is,{n:5,done:G[4],current:Q===4,title:"MCP entry in the repository (optional)",children:a.jsxs("label",{className:"check-label compact",children:[a.jsx("input",{type:"checkbox",checked:x,onChange:U=>y(U.target.checked)}),"The agent adds it as part of step 1. To do it by hand, copy it from the ",a.jsx("button",{type:"button",className:"text-button",onClick:_,children:"Coding agent brief"}),"."]})}),a.jsxs(Is,{n:6,done:G[5],current:Q===5,title:"Connect a coding agent (optional)",children:[a.jsxs("p",{className:"step-why",children:[a.jsx("strong",{children:"Why connect one:"})," when this application keeps failing, aiport can show you more than the error. With a coding agent and the repository, ",a.jsx("strong",{children:"Explain"})," reads the application's code and tells you the root cause, and ",a.jsx("strong",{children:"Fix it"})," makes the change on a new branch for you to review. Nothing runs until you click, Explain only reads, and aiport never commits to your branch, merges or pushes. Without one, aiport still explains from the evidence alone, more briefly."]}),a.jsxs("p",{children:["The coding agent aiport may start, on your click, to explain an Issue of this application or fix it in its own branch, and the repository it works in. Choose among the agents installed on this machine",v.default_agent?a.jsx(a.Fragment,{children:", or keep the default from Settings"}):a.jsx(a.Fragment,{children:"; a default for every application is in Settings"}),"."]}),H?a.jsx(Aw,{o:H,agents:v,onDone:E}):"Once paired: the repository path the agent reported while pairing, as you confirmed it, appears here."]})]}),Q<0&&a.jsxs("p",{className:"notice",role:"status",children:[a.jsx(ug,{size:16,"aria-hidden":!0})," This application is onboarded."]})]})]})}function Um({id:o,title:c,short:u,icon:r,brief:h,link:m,open:g,setOpen:v,snippets:b,picker:C,onSnippetCopied:x}){const[y,w]=f.useState("");return f.useEffect(()=>{m&&fetch(Ml(m.url),{credentials:"same-origin"}).then(A=>A.ok?A.text():"").then(w).catch(()=>{})},[m?.url,h]),a.jsxs("section",{className:"tile"+(g?" open":""),id:"tile-"+o,"aria-labelledby":"tile-heading-"+o,children:[a.jsxs("div",{className:"tile-head",children:[a.jsx("span",{className:"group-icon","aria-hidden":!0,children:r}),a.jsxs("div",{className:"tile-title",children:[a.jsx("h3",{id:"tile-heading-"+o,children:a.jsxs("button",{type:"button",className:"group-toggle","aria-expanded":g,"aria-controls":"tile-body-"+o,onClick:()=>v(!g),children:[a.jsx(un,{size:15,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),c]})}),a.jsx("p",{className:"tile-audience",title:m?.audience,children:u})]})]}),a.jsxs("div",{className:"tile-actions",children:[a.jsx(Ka,{text:y,label:"Copy whole brief"}),m&&a.jsxs("a",{className:"tile-url",href:Ml(m.url),target:"_blank",rel:"noreferrer",title:"Open the brief",children:[a.jsx("code",{children:Ml(m.url)}),a.jsx(Sd,{size:12,"aria-hidden":!0})]})]}),g&&a.jsxs("div",{className:"brief tile-body",id:"tile-body-"+o,children:[m?.audience&&a.jsx("p",{className:"muted-small",children:m.audience}),a.jsx("p",{children:a.jsx(In,{text:h.intro})}),a.jsx("ol",{children:h.steps.map((A,E)=>a.jsx("li",{children:a.jsx(In,{text:A})},E))}),a.jsxs("div",{className:"snippets-heading",children:[a.jsx("h4",{children:"Snippets"}),C]}),b.map(A=>a.jsx(Mw,{s:A,onCopied:()=>x?.(A)},A.id)),a.jsx("p",{className:"muted-small",children:a.jsx(In,{text:h.note})})]})]})}function Xw({origins:o,targets:c,providers:u=[],tick:r,onChange:h,onManual:m,onChanges:g}){const[v,b]=f.useState(null),[C,x]=f.useState(""),[y,w]=f.useState(()=>["python","typescript","curl"].includes(Fi(bd))?Fi(bd):"python"),[A,E]=f.useState(Ys),[_,N]=f.useState(()=>zd(Ys)),[D,H]=mv(),[G,Q]=hw(r),Y=qw(o.map(de=>de.id),r),{issues:S,explain:U,reload:Z}=Dw(r),[$,oe]=kv(r),be=Sv(r),re=gw(r),[J]=yv(r),[ve,W]=f.useState(null),[F,ne]=f.useState(null),[le,je]=f.useState(!1),[Ae,Qe]=f.useState(!1),Ze=f.useRef(null),dt=f.useRef(new Set(o.map(de=>de.id))),j=D.length>0;f.useEffect(()=>{N(zd(A))},[A]);const I=de=>{Fi(Tv+A,de?"1":"0"),N(de)};f.useEffect(()=>{const de=o.find(Fe=>!dt.current.has(Fe.id));o.forEach(Fe=>dt.current.add(Fe.id)),de&&A===Ys&&E(de.id)},[o]);const ee=f.useRef(!1);if(f.useEffect(()=>{j&&!ee.current&&W(!0),ee.current=j},[j]),f.useEffect(()=>{fetch("/_connect",{credentials:"same-origin"}).then(de=>{if(!de.ok)throw new Error;return de.json()}).then(b).catch(()=>x("Could not load the instructions from /_connect."))},[r]),C)return a.jsx("p",{className:"auth-error",role:"alert",children:C});if(!v)return a.jsx("p",{className:"muted",children:"Loading…"});const ie=ve??(o.length===0||j),fe=de=>{E(de),W(!0),requestAnimationFrame(()=>Ze.current?.scrollIntoView({behavior:"smooth",block:"nearest"}))},ye=de=>{E(de),W(!0),requestAnimationFrame(()=>document.getElementById("agent-"+de)?.focus())},xe=()=>{H(),Q(),h()},ae=de=>{history.replaceState(null,"","#"+Xd(de.id)),g?.()},he=de=>{w(de),Fi(bd,de)},ut=de=>v.briefs.find(Fe=>Fe.for===de),Qt=v.integration,B=v.agent_integration,Re=Qt.snippets.filter(de=>de.sdk&&de.language===y).sort((de,Fe)=>Number(Fe.available)-Number(de.available)),Ie=a.jsx("div",{className:"segmented languages",role:"radiogroup","aria-label":"Language",children:zw.map(([de,Fe])=>a.jsx("button",{type:"button",role:"radio","aria-checked":y===de,className:y===de?"on":"",onClick:()=>he(de),children:Fe},de))}),we=F??o.length===0,st=()=>{ne(!0),Qe(!0),requestAnimationFrame(()=>document.getElementById("tile-agent")?.scrollIntoView({behavior:"smooth",block:"start"}))};return a.jsxs("div",{className:"integrate-layout",children:[a.jsxs("section",{className:"panel wizard instructions-panel","aria-labelledby":"strip-heading",children:[a.jsxs("div",{className:"panel-heading wizard-heading",children:[a.jsxs("h2",{id:"strip-heading",children:[a.jsxs("button",{type:"button",className:"group-toggle","aria-expanded":we,"aria-controls":"strip-body",onClick:()=>ne(!we),children:[a.jsx(un,{size:16,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),"Instructions"]}),a.jsx(wt,{section:"The instructions"})]}),a.jsx("span",{className:"muted-small",children:"Communication standard · Application brief · Coding agent brief"})]}),we&&a.jsxs("div",{className:"strip-body",id:"strip-body",children:[a.jsxs("p",{className:"muted-small strip-intro",children:["What the coding agent reads, written by aiport from this gateway's address and protocols (",v.protocols.map(de=>de.name+(de.available?" configured":" not configured")).join(", "),"). ",a.jsx("button",{type:"button",className:"text-button",onClick:m,children:"Read the manual"})]}),a.jsxs("article",{className:"standard-card","aria-labelledby":"standard-heading",children:[a.jsxs("div",{className:"standard-head",children:[a.jsx(kx,{size:17,strokeWidth:1.75,"aria-hidden":!0}),a.jsxs("div",{children:[a.jsx("h3",{id:"standard-heading",children:"Communication standard"}),a.jsxs("p",{className:"muted-small",children:["Every rule an integration follows, with a checklist: the token in ",a.jsx("code",{children:".aiport"}),", Levels and task names, ",a.jsx("code",{children:"X-Aiport-Effort"}),", ",a.jsx("code",{children:"X-Aiport-Module"})," naming the module of every call, a thumbs up and down on every answer sent to aiport, and the error fixes. Paste it to the coding agent of a new application, or of one already connected, to bring it in line."]})]})]}),a.jsx("blockquote",{className:"agent-prompt",children:v.standard_prompt}),a.jsxs("div",{className:"button-row",children:[a.jsx(Ka,{text:v.standard_prompt,label:"Copy instruction"}),a.jsx(Ka,{text:v.standard,label:"Copy URL only"}),a.jsxs("a",{className:"text-button",href:"/_connect?format=standard",target:"_blank",rel:"noreferrer",children:["Read the standard ",a.jsx(Sd,{size:12,"aria-hidden":!0})]})]})]}),a.jsx(Um,{id:"application",title:"Application brief",short:"For the agent wiring the application's LLM calls through aiport.",icon:a.jsx(Vx,{size:17,strokeWidth:1.75}),brief:Qt,link:ut("application"),open:le,setOpen:je,picker:Ie,snippets:[...Qt.snippets.filter(de=>!de.sdk),...Re]}),a.jsx(Um,{id:"agent",title:"Coding agent brief",short:"For agents working in its repository later: MCP entry, AGENTS.md, Change requests.",icon:a.jsx(Ud,{size:17,strokeWidth:1.75}),brief:B,link:ut("agent"),open:Ae,setOpen:Qe,snippets:B.snippets,onSnippetCopied:de=>{de.id==="mcp"&&I(!0)}}),a.jsxs("p",{className:"muted-small strip-foot",children:["Reference: ",a.jsxs("a",{href:"/_connect?format=markdown",target:"_blank",rel:"noreferrer",children:[a.jsx("code",{children:"/_connect?format=markdown"})," ",a.jsx(Sd,{size:12,"aria-hidden":!0})]})]})]})]}),a.jsxs("div",{className:"integrate-main",children:[a.jsx(Bw,{origins:o,targets:c,providers:u,pairings:D,rejected:G,calls:Y,issues:S,guidance:be,explain:U,onExplained:Z,agents:J,levels:$,onLevels:()=>{oe(),h()},app:A,adapting:v.adapting,onSelect:E,onAgent:ye,onNew:()=>fe(Ys),onDone:xe,onForgot:Q,onIssue:ae,missing:re}),a.jsx("div",{ref:Ze,children:a.jsx(Yw,{origins:o,app:A,setApp:E,pairings:D,seen:A?Y.at[A]:null,off:Y.off,adapting:v.adapting,agents:J,prompt:v.agent_prompt,docs:v.docs,mcp:_,setMcp:I,open:ie,setOpen:W,onDone:xe,onAgentBrief:st})})]})]})}function Ev(o){const c=r=>r.connection+"/"+r.model+(r.level&&r.protocol?" on "+(r.protocol==="openai"?"OpenAI":"Anthropic"):""),u=r=>r==="binding"?"Binding":r==="openai"?"OpenAI default":r==="anthropic"?"Anthropic default":r;return[...o.removed.map(r=>"“"+r.name+"” removed (was "+c(r)+")"),...o.moved.map(r=>"“"+r.name+"” now "+c(r.to)+" (was "+c(r.from)+")"),...(o.added||[]).map(r=>"“"+r.name+"” added ("+c(r)+")"),...o.routing.map(r=>u(r.of)+" now "+(r.after||"none")+(r.before?" (was "+r.before+")":"")),...(o.connections_removed||[]).map(r=>"no longer reaches "+r),...(o.connections_added||[]).map(r=>"now reaches "+r)]}const Pi=o=>o.cause==="catalog_changed";function Vw({issue:o}){const c=o.report,u=o.evidence;return a.jsxs("div",{className:"evidence",children:["You reported ",u.count===1?"this failed request":u.count+" failed requests of this kind"," from Activity",u.count>1?" between "+Rt(u.from_ms)+" and "+Rt(u.to_ms):" at "+Rt(c.at_ms),": ",a.jsx("code",{children:c.request_id}),c.module?a.jsxs(a.Fragment,{children:[", module ",a.jsx("code",{children:c.module})]}):null,c.model?", served by "+c.model+(c.connection?" on "+c.connection:""):"",".",c.note&&a.jsx("blockquote",{className:"report-note",children:c.note})]})}function Ov({issue:o}){const c=o.catalog_change;if(!c)return null;const u=Ev({removed:c.removed,moved:c.moved,routing:c.routing});return a.jsxs(a.Fragment,{children:[a.jsxs("div",{className:"evidence",children:["Catalog change (Change ",c.change_id,", ",Rt(c.at_ms),"): ",u.join("; "),". ",o.application," was told; nothing waits for you."]}),a.jsxs("div",{className:"evidence",children:["Still available: ",c.available.length?c.available.map(r=>"“"+r.name+"”").join(", "):"no task names",c.connections.length?", and the models of "+c.connections.join(", "):"","."]})]})}const Qw=["ready","stopped_waiting","timed_out"];function Zw({approval:o,onAnswered:c}){const[u,r]=f.useState(""),[h,m]=f.useState(""),g=async b=>{r(b),m("");try{await Vt("POST","approvals/"+encodeURIComponent(o.id),{answer:b}),c()}catch(C){m(C.message)}finally{r("")}},v=Math.max(0,Math.round((o.expires_at_ms-o.asked_at_ms)/6e4));return a.jsxs("div",{className:"approval",role:"group","aria-label":"Command waiting for your answer",children:[a.jsxs("p",{children:[a.jsx("strong",{children:"The agent asks to run a command"})," in its worktree (the whole command; invisible characters are shown as escapes such as ","\\u202e","):"]}),a.jsx("pre",{className:"approval-command",tabIndex:0,"aria-label":"The whole command",children:o.command}),a.jsxs("div",{className:"actions",children:[a.jsx("button",{type:"button",className:"primary small",disabled:!!u,onClick:()=>g("allow_once"),children:"Allow once"}),a.jsx("button",{type:"button",className:"secondary small",disabled:!!u,onClick:()=>g("allow_run"),title:"Allow this exact command again in this run, until the agent changes a file",children:"Allow until the next edit"}),a.jsx("button",{type:"button",className:"secondary small",disabled:!!u,onClick:()=>g("deny"),children:"Deny"}),a.jsxs("span",{className:"caption",children:["Unanswered for ",v," min, the run stops and keeps what it changed for your review."]})]}),h&&a.jsx("p",{className:"auth-error",role:"alert",children:h})]})}const zv=o=>o.explain?{...o.explain,by:o.explain_by,fix:o.fix_by}:null,Kw={claude:"Claude Code",codex:"Codex",gemini:"Gemini CLI"},Zi=o=>o.source==="coding_agent"?(o.agent_name||Kw[o.agent||""]||o.agent)+", after reading the code":o.connection+" ("+o.model+")",Fw=o=>o?o.through_aiport?"Model calls go through aiport to "+(o.connection||"the Anthropic protocol default")+(o.model?" ("+o.model+")":"")+", under the Application aiport.":o.note||"The agent uses its own login.":"";function Rv({run:o,steps:c,busy:u,branch:r}){const[h,m]=f.useState(!1),[g,v]=f.useState(""),b=async()=>{m(!0),v("");const x=await fetch("/api/v1/runs/"+encodeURIComponent(o.id)+"/stop",{method:"POST",credentials:"same-origin",headers:{"X-Aiport-CSRF":"1"}}).catch(()=>null);(!x||!x.ok&&x.status!==404)&&(v("Could not stop the run."),m(!1))},C=c.slice(-30);return a.jsxs("div",{className:"run-panel","aria-label":"Agent run of "+o.agent_name,children:[a.jsxs("div",{className:"run-head",children:[a.jsx("strong",{children:o.agent_name})," ",a.jsxs("span",{className:"caption",children:[o.mode==="fix"?"changes the code in its own worktree"+(r?" on "+r:"")+" · commands wait for your approval":"reads the code, read-only"," · stops after ",Math.round(o.limit_ms/6e4)," min"]}),u&&a.jsx("button",{type:"button",className:"secondary small",disabled:h,onClick:b,children:h?"Stopping…":"Stop"})]}),a.jsx("p",{className:"caption"+(o.model_calls.through_aiport?"":" run-own-login"),children:Fw(o.model_calls)}),o.session_mode&&a.jsxs("p",{className:"caption run-mode",children:["Session mode: ",o.session_mode]}),!!o.warnings?.length&&a.jsxs("div",{className:"notice-inline",role:"note",children:[a.jsx("strong",{children:"The repository's agent settings allow some things without asking."})," aiport never sees those, so it cannot ask you about them:",a.jsx("ul",{className:"run-warnings",children:o.warnings.map((x,y)=>a.jsx("li",{children:x},y))})]}),C.length>0&&a.jsx("ol",{className:"run-steps","aria-live":"polite",children:C.map((x,y)=>a.jsx("li",{className:"run-step run-"+x.kind+(x.verdict?" run-"+x.verdict:""),children:x.kind==="permission"?a.jsxs(a.Fragment,{children:[x.verdict==="allowed"?"Allowed":"Denied",": ",x.text,x.reason?a.jsxs("small",{children:[" · ",x.reason]}):null]}):x.kind==="tool"?a.jsxs(a.Fragment,{children:[x.text,x.status&&x.status!=="completed"?a.jsxs("small",{children:[" · ",x.status]}):null]}):x.kind==="thought"?a.jsx("em",{children:x.text}):x.text},y))}),g&&a.jsx("p",{className:"auth-error",role:"alert",children:g})]})}function Mv({issue:o,explain:c,onExplained:u}){const r=c?.by?.[o.application]||c,[h,m]=f.useState(""),[g,v]=f.useState(!1),[b,C]=f.useState(""),[x,y]=f.useState(""),[w,A]=f.useState(null),[E,_]=f.useState([]),N=async Q=>{let Y=!1;for(;Q();){const S=await fetch("/api/v1/issues/"+encodeURIComponent(o.id)+"/run",{credentials:"same-origin"}).then(Z=>Z.ok?Z.json():null).catch(()=>null);if(!Q()||!S?.run||S.run.mode!=="explain"||!Y&&S.run.state!=="running")break;Y=!0;const U=S.log||[];if(v(!0),A(S.run),y(Zi({source:"coding_agent",agent:S.run.agent,agent_name:S.run.agent_name})),m(U.filter(Z=>Z.kind==="text").map(Z=>Z.text).join("")),_(U.filter(Z=>Z.kind!=="text"&&Z.kind!=="status")),S.run.state!=="running"){v(!1),S.run.state==="done"?u():C("The Agent run ended ("+S.run.state+")"+(S.run.error?": "+S.run.error:"")+"; nothing was saved.");return}await new Promise(Z=>setTimeout(Z,1500))}Y&&v(!1)};f.useEffect(()=>{if(r?.source!=="coding_agent")return;let Q=!0;return N(()=>Q),()=>{Q=!1}},[o.id,r?.source]);const D=async()=>{v(!0),C(""),m(""),y(""),A(null),_([]);try{const Q=await fetch("/api/v1/issues/"+encodeURIComponent(o.id)+"/explain",{method:"POST",credentials:"same-origin",headers:{"X-Aiport-CSRF":"1"}});if(!Q.ok||!Q.body){const $=await Q.json().catch(()=>({}));throw new Error($.error?.message||"Explain failed ("+Q.status+").")}const Y=Q.body.pipeThrough(new TextDecoderStream).getReader();let S="",U=!1,Z="";for(;!U;){const $=await Y.read();if($.done)break;S+=$.value;let oe;for(;(oe=S.indexOf(`
`))>=0;){const be=S.slice(0,oe).trim();if(S=S.slice(oe+1),!be)continue;const re=JSON.parse(be);re.type==="start"?y(Zi({...re.by,agent_name:r?.agent_name})):re.type==="run"?A(re.run):re.type==="step"?_(J=>[...J,re.step]):re.type==="text"?m(J=>J+re.text):re.type==="done"?U=!0:re.type==="error"&&(Z=re.error?.message||"Explain failed.")}}if(Z)throw new Error(Z);if(!U&&r?.source==="coding_agent"){v(!1),await N(()=>!0);return}if(!U)throw new Error("The explanation broke off; nothing was saved.");u()}catch(Q){C(Q.message),Q instanceof TypeError&&r?.source==="coding_agent"&&(C(""),await N(()=>!0))}finally{v(!1)}},H=o.explanation,G=g||h&&b?h:H?.text||"";return a.jsxs("div",{className:"explanation",children:[G&&a.jsxs("div",{className:"explanation-text","aria-live":"polite",children:[a.jsx("strong",{children:"Root cause"}),(g?x:H&&Zi(H))&&a.jsxs("small",{children:[" · written by ",g?x:Zi(H),!g&&H?.source==="connection"?" from the evidence only":""]}),a.jsx("p",{children:G})]}),w&&a.jsx(Rv,{run:w,steps:E,busy:g}),g&&!h&&!w&&a.jsxs("p",{className:"caption",role:"status",children:["Asking ",x||"the model","…"]}),b&&a.jsxs("div",{className:"notice error",role:"alert",children:[b,H?" The earlier explanation stays.":""]}),a.jsxs("div",{className:"actions",children:[a.jsxs("button",{type:"button",className:"secondary",disabled:g||!r?.available,onClick:D,title:r?.available?"Explain the root cause with "+Zi(r):r?.hint,children:[a.jsx(Yx,{size:14,"aria-hidden":!0})," ",g?"Explaining…":H?"Explain again":"Explain"]}),r&&!r.available&&r.hint&&a.jsx("span",{className:"caption",children:r.hint})]})]})}function Jw({diff:o,truncated:c}){const u=o.split(`
`);return a.jsxs("pre",{className:"diff","aria-label":"Diff",children:[u.map((r,h)=>a.jsx("span",{className:r.startsWith("+++")||r.startsWith("---")?"diff-file":r.startsWith("+")?"diff-add":r.startsWith("-")?"diff-del":r.startsWith("@@")?"diff-hunk":void 0,children:r+`
`},h)),c&&a.jsx("span",{className:"diff-hunk",children:"… the diff is longer than aiport shows; read the rest on the branch."})]})}function qv({issue:o,explain:c,onChanged:u}){const r=c?.fix?.[o.application],h=o.branch_fix,[m,g]=f.useState(""),[v,b]=f.useState(""),[C,x]=f.useState(null),[y,w]=f.useState([]),[A,E]=f.useState(null),[_,N]=f.useState([]),D=h?.state==="running",H=!!h&&Qw.includes(h.state);f.useEffect(()=>{if(!D)return;let S=!1;return(async()=>{for(;!S;){const Z=await fetch("/api/v1/issues/"+encodeURIComponent(o.id)+"/run",{credentials:"same-origin"}).then($=>$.ok?$.json():null).catch(()=>null);if(S)return;Z?.run&&(x(Z.run),w((Z.log||[]).filter($=>$.kind!=="status")),N(Z.approvals||[])),Z?.run&&Z.run.state!=="running"&&u(),await new Promise($=>setTimeout($,1500))}})(),()=>{S=!0,N([])}},[o.id,D]),f.useEffect(()=>{E(null),!(!h||!H&&h.state!=="kept")&&fetch("/api/v1/issues/"+encodeURIComponent(o.id)+"/fix",{credentials:"same-origin"}).then(S=>S.ok?S.json():null).then(S=>S&&E({diff:S.diff||"",truncated:!!S.truncated})).catch(()=>{})},[o.id,h?.state,h?.run]);const G=async S=>{g(S),b("");try{await Vt("POST","issues/"+encodeURIComponent(o.id)+"/"+S),x(null),w([]),u()}catch(U){b(U.message)}finally{g("")}},Q=!r?.causes||r.causes.includes(o.cause),Y=r?.available&&Q&&o.state==="open"&&!D&&!H;return!h&&!r?.available?r?.hint?a.jsx("p",{className:"caption fix-hint",children:r.hint}):null:!h&&!Q?r?.cause_hint?a.jsx("p",{className:"caption fix-hint",children:r.cause_hint}):null:a.jsxs("div",{className:"branch-fix",children:[D&&C&&a.jsx(Rv,{run:C,steps:y,busy:!0,branch:h?.branch}),D&&_.map(S=>a.jsx(Zw,{approval:S,onAnswered:()=>{N(U=>U.filter(Z=>Z.id!==S.id)),u()}},S.id)),D&&!C&&a.jsxs("p",{className:"caption",role:"status",children:["Starting ",r?.agent_name||h?.agent," in a worktree on ",h?.branch,"…"]}),H&&a.jsxs("div",{className:"fix-review",children:[h.state!=="ready"&&h.note&&a.jsx("p",{className:"caption",children:h.note}),a.jsxs("p",{children:[a.jsx("strong",{children:"Review the change"})," on branch ",a.jsx("code",{children:h.branch}),": ",h.summary,". Nothing is committed yet; your branch and working copy are untouched."]}),A&&a.jsx(Jw,{diff:A.diff,truncated:A.truncated}),a.jsxs("div",{className:"actions",children:[a.jsx("button",{type:"button",className:"primary",disabled:!!m,onClick:()=>G("fix/keep"),children:m==="fix/keep"?"Keeping…":"Keep the branch"}),a.jsx("button",{type:"button",className:"secondary",disabled:!!m,onClick:()=>G("fix/discard"),children:m==="fix/discard"?"Discarding…":"Discard"}),a.jsxs("span",{className:"caption",children:["Keep commits it on ",h.branch," as aiport; merging is yours. Discard deletes the worktree and the branch."]})]})]}),h?.state==="kept"&&a.jsxs("p",{className:"caption",children:["Kept on branch ",a.jsx("code",{children:h.branch}),h.commit?" (commit "+h.commit.slice(0,12)+")":"",": ",h.summary,". aiport never merges or pushes: merge it when you are ready; the Issue resolves when ",o.application,"’s calls succeed."]}),h&&(h.state==="discarded"||h.state==="stopped"||h.state==="empty")&&a.jsxs("p",{className:"caption",children:["Last Fix it run: ",h.state==="discarded"?"discarded; the worktree and the branch were removed.":h.note]}),v&&a.jsx("div",{className:"notice error",role:"alert",children:v}),Y&&a.jsxs("div",{className:"actions",children:[a.jsx("button",{type:"button",className:"secondary",disabled:!!m,onClick:()=>G("fix"),title:"Fix it with "+(r?.agent_name||r?.agent)+" in its own worktree; stops after "+Math.round((r?.limit_ms||0)/6e4)+" min",children:m==="fix"?"Starting…":h?"Fix it again":"Fix it"}),a.jsxs("span",{className:"caption",children:[r?.agent_name," edits a new branch in its own worktree; you approve each command it runs; you review the diff."]})]})]})}const Dv="aiport (automatic)";function Ww({issue:o,targets:c,explain:u,onDecided:r,onExplained:h}){const m=o.fix,g=o.evidence,v=o.cause==="failing_connection",b=o.cause==="level_unmapped",[C,x]=f.useState(m?.proposal.target||""),[y,w]=f.useState(m?.proposal.model||(b?"":o.subject)),[A,E]=f.useState(""),[_,N]=f.useState(""),[D,H]=f.useState(null),G=m?.state==="applied"&&m.by===Dv&&!!m.change_id,Q=async()=>{E("revert"),N(""),H(null);try{await Vt("POST","changes/"+encodeURIComponent(m.change_id)+"/revert"),r("Reverted the automatic fix. aiport will not apply it automatically again for "+o.application+"; the Issue waits for your decision.")}catch(S){N(S.message)}finally{E("")}},Y=async S=>{E(S),N(""),H(null);try{await Vt("POST","issues/"+encodeURIComponent(o.id),S!=="accept"?{action:S}:v?{action:S,target:C}:{action:S,target:C,model:y}),r(S==="accept"?"Approved for only "+o.application+(v?": bound to "+C:b?": the Level “"+o.subject+"” → "+C+"/"+y:": the alias “"+o.subject+"” → "+C+"/"+y)+". The Issue resolves itself when its calls succeed; you can revert the Change below.":"Dismissed. aiport stays quiet about this Issue for 7 days unless the evidence doubles.")}catch(U){const Z=U;Z.code==="guardrail_blocked"?H({extension:Z.extension||"a guardrail",reason:Z.reason||""}):N(Z.message)}finally{E("")}};return a.jsxs("div",{className:"decision issue",id:Xd(o.id),tabIndex:-1,children:[a.jsxs("div",{className:"decision-head",children:[a.jsx("h3",{className:"need",children:o.summary}),a.jsxs("div",{className:"decision-meta",children:[a.jsx("span",{className:"badge"+(o.state==="resolving"?" good":" warn"),children:o.state==="resolving"?"Resolving":"Open"}),a.jsx("span",{className:"badge",children:o.application})]})]}),Pi(o)?a.jsx(Ov,{issue:o}):o.report?a.jsx(Vw,{issue:o}):a.jsxs("div",{className:"evidence",children:["Evidence: ",Rd(g.count)," ",g.status,g.error?" "+g.error:""," between ",Rt(g.from_ms)," and ",Rt(g.to_ms)," (looking back ",g.window_minutes," minutes)",g.connections?.length?", sent to "+g.connections.join(", "):"",".",g.examples?.length?" Example requests: "+g.examples.slice(0,3).join(", ")+".":"",o.reopened?" Came back "+Rd(o.reopened)+".":""]}),o.diagnosis?.summary&&a.jsxs("div",{className:"evidence",children:["Diagnosis: ",o.diagnosis.summary]}),G&&a.jsxs("div",{className:"evidence",children:[a.jsx("strong",{children:"aiport fixed this automatically:"})," the alias “",m.proposal.alias,"” → ",m.proposal.target,"/",m.proposal.model," for only ",o.application,". ",a.jsxs("button",{className:"text-button",disabled:!!A,onClick:Q,"aria-label":"Revert the automatic fix for "+o.application,children:[a.jsx(Id,{size:14,"aria-hidden":!0})," ",A==="revert"?"Reverting…":"Revert"]})]}),o.state==="resolving"&&!G&&a.jsxs("div",{className:"evidence",children:[o.resolving_by==="aiport_fix"?"The aiport fix is applied"+(m?.by?" (by "+m.by+")":""):"The Application says it fixed this",": aiport marks the Issue resolved when ",o.application,"’s calls ",v?"succeed again":"for "+o.subject+" succeed","."]}),_&&a.jsx("div",{className:"notice error",role:"alert",children:_}),D&&a.jsxs("div",{className:"notice error",role:"alert",children:[a.jsxs("strong",{children:["Blocked by the guardrail ",D.extension,":"]})," ",D.reason," Nothing was changed; this Issue stays open."]}),m?.state==="proposed"&&a.jsxs("form",{className:"decision-form",onSubmit:S=>{S.preventDefault(),Y("accept")},children:[v?a.jsxs("label",{children:["Fix in aiport: bind only ",o.application," to",a.jsxs("select",{value:C,onChange:S=>x(S.target.value),children:[!C&&a.jsx("option",{value:"",children:"Choose a working Connection"}),c.filter(S=>S!==o.subject).map(S=>a.jsx("option",{children:S},S))]})]}):a.jsxs(a.Fragment,{children:[a.jsxs("label",{children:[b?"Fix in aiport: map the Level ("+(o.protocol==="anthropic"?"Anthropic":"OpenAI")+" protocol) for only "+o.application:"Fix in aiport: alias for only "+o.application,a.jsx("input",{value:o.subject,readOnly:!0,spellCheck:!1})]}),a.jsxs("label",{children:["Connection",a.jsxs("select",{value:C,onChange:S=>x(S.target.value),children:[!C&&a.jsx("option",{value:"",children:"Choose a Connection"}),c.map(S=>a.jsx("option",{children:S},S))]})]}),a.jsxs("label",{children:["Model to use",a.jsx("input",{value:y,onChange:S=>w(S.target.value),required:!0,maxLength:256,spellCheck:!1,autoComplete:"off",placeholder:b?"A model the Connection serves":void 0})]})]}),a.jsxs("div",{className:"actions",children:[a.jsx("button",{className:"primary",disabled:!!A||!C||!v&&!y.trim(),children:A==="accept"?"Approving…":"Approve"}),a.jsx("button",{type:"button",className:"secondary",disabled:!!A,onClick:()=>Y("dismiss"),children:A==="dismiss"?"Dismissing…":"Dismiss"})]})]}),m?.state!=="proposed"&&a.jsx("div",{className:"actions",children:a.jsx("button",{type:"button",className:"secondary",disabled:!!A,onClick:()=>Y("dismiss"),children:A==="dismiss"?"Dismissing…":"Dismiss"})}),!Pi(o)&&a.jsx(Mv,{issue:o,explain:u,onExplained:h}),a.jsx(qv,{issue:o,explain:u,onChanged:h}),a.jsx(Lv,{issue:o})]})}const Xd=o=>"issue-"+o;function Lv({issue:o}){return o.application_fix?a.jsxs("div",{className:"application-fix",children:[a.jsx(Ka,{text:o.application_fix,label:"Copy fix for the Coding agent"}),a.jsxs("details",{children:[a.jsxs("summary",{children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),"Read the fix for ",o.application,"’s Coding agent"]}),a.jsx("pre",{children:o.application_fix})]})]}):null}const Uv=o=>{const c=o.match(/^.*?[.!?](?=\s|$)/s);return(c?c[0]:o).trim()},Rd=o=>o+" "+(o===1?"time":"times"),Wi=(o,c)=>o==="every_application"?"every Application":"only "+c;async function Vt(o,c,u){const r=await fetch("/api/v1/"+c,{method:o,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:u===void 0?void 0:JSON.stringify(u)}),h=await r.json().catch(()=>({}));if(!r.ok)throw Object.assign(new Error(h.error?.message||"Request failed ("+r.status+")."),{code:h.error?.code,extension:h.error?.extension,reason:h.error?.reason});return h}const Rt=o=>new Date(o).toLocaleString();function Pw({item:o,targets:c,onDecided:u,onOpenApplications:r}){const h=o.source==="request",m=o.kind==="application_alias",g=!h&&o.kind==="bind_application"?o.subject:void 0,v=c.filter(W=>W!==g),[b,C]=f.useState(o.proposal.target&&v.includes(o.proposal.target)?o.proposal.target:v[0]||""),[x,y]=f.useState(o.proposal.model||""),[w,A]=f.useState(o.proposal.alias||""),[E,_]=f.useState(o.scope||"application"),[N,D]=f.useState(""),[H,G]=f.useState(""),[Q,Y]=f.useState(""),[S,U]=f.useState(null),Z=h?"reject":"dismiss",$=async W=>{G(W),Y(""),U(null);try{await Vt("POST","decisions/"+encodeURIComponent(o.id),W==="accept"?{action:W,scope:E,alias:m?w:void 0,target:b,model:m?x:void 0,note:N}:{action:W,note:N});const F=Wi(E,o.application);u(W==="accept"?(m?"Approved for "+F+": the alias “"+w+"” → "+b+"/"+x+".":E==="every_application"?"Approved: "+b+" is now the default Connection for its protocol.":"Approved: "+o.application+" now sends its calls to "+b+".")+" You can revert it below.":W==="reject"?"Rejected. "+((o.asked_by?.length||0)>1?"Each Application that asked sees":o.application+" sees")+" your note when it lists its requests.":"Dismissed. aiport stays quiet about this for 7 days unless the evidence doubles.")}catch(F){const ne=F;ne.code==="guardrail_blocked"?U({extension:ne.extension||"a guardrail",reason:ne.reason||""}):Y(ne.message)}finally{G("")}},oe=o.evidence,be=o.asked_by||[],re=be.length>1,J=!!b&&(!m||!!w.trim()&&!!x.trim()),ve=o.kind==="application_alias"||o.kind==="bind_application"||!!o.informational;return a.jsxs("div",{className:"decision",children:[a.jsxs("div",{className:"decision-head",children:[a.jsx("h3",{className:"need",children:o.need}),a.jsxs("div",{className:"decision-meta",children:[a.jsx("span",{className:"badge",children:h?re?"Asked by "+be.length+" Applications":"Request from "+o.application:"Suggestion"}),a.jsxs("span",{className:"badge",children:["For ",Wi(o.scope,o.application)]}),!ve&&a.jsx("span",{className:"badge warn",children:"Not supported yet"})]})]}),oe&&o.kind==="pair_application"&&a.jsxs("div",{className:"evidence",children:["Evidence: ",oe.count," refused requests between ",Rt(oe.from_ms)," and ",Rt(oe.to_ms)," (looking back ",oe.window_minutes," minutes). Only counts are kept, never the token."]}),oe&&o.kind!=="pair_application"&&a.jsxs("div",{className:"evidence",children:["Evidence: ",oe.count," requests between ",Rt(oe.from_ms)," and ",Rt(oe.to_ms)," (looking back ",oe.window_minutes," minutes). Only request metadata was used."]}),h&&o.details&&a.jsxs("p",{className:"decision-ask",children:[a.jsx("span",{children:"Asks for"})," ",Uv(o.details)]}),a.jsx(Hv,{item:o,targets:v}),h&&a.jsxs("details",{className:"decision-more",children:[a.jsxs("summary",{children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),"Details and reasons"]}),!re&&a.jsxs("dl",{className:"request-text",children:[a.jsx("dt",{children:"Details"}),a.jsx("dd",{children:o.details}),a.jsx("dt",{children:"Why"}),a.jsx("dd",{children:o.why}),a.jsxs("dt",{children:["Why ",Wi(o.scope,o.application)]}),a.jsx("dd",{children:o.scope_case}),o.filed_by&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Filed by (as stated)"}),a.jsx("dd",{children:o.filed_by})]}),!!o.repeats&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Asked again"}),a.jsxs("dd",{children:[o.repeats," more ",o.repeats===1?"time":"times"]})]})]}),re&&a.jsxs(a.Fragment,{children:[a.jsxs("dl",{className:"request-text",children:[a.jsx("dt",{children:"Details"}),a.jsx("dd",{children:o.details})]}),a.jsx("ul",{className:"askers","aria-label":"Each Application's argument",children:be.map(W=>a.jsxs("li",{children:[a.jsx("strong",{children:W.application}),W.filed_by?" (filed by "+W.filed_by+", as stated)":"",W.repeats?" · asked again "+Rd(W.repeats):"",a.jsxs("dl",{className:"request-text",children:[a.jsx("dt",{children:"Why"}),a.jsx("dd",{children:W.why}),a.jsx("dt",{children:"Argument for every Application"}),a.jsx("dd",{children:W.scope_case})]})]},W.application))})]})]}),Q&&a.jsx("div",{className:"notice error",role:"alert",children:Q}),S&&a.jsxs("div",{className:"notice error",role:"alert",children:[a.jsxs("strong",{children:["Blocked by the guardrail ",S.extension,":"]})," ",S.reason," Nothing was changed; this item still waits. Reject it, or change what the guardrail checks and approve again."]}),(m||o.kind==="bind_application")&&a.jsxs("form",{className:"decision-form",onSubmit:W=>{W.preventDefault(),$("accept")},children:[a.jsxs("label",{children:["Scope",a.jsxs("select",{value:E,onChange:W=>_(W.target.value),children:[a.jsxs("option",{value:"application",children:["Only ",o.application]}),a.jsx("option",{value:"every_application",children:"Every Application"})]})]}),m&&a.jsxs("label",{children:["Alias",a.jsx("input",{value:w,onChange:W=>A(W.target.value),readOnly:!h,required:!0,maxLength:256,spellCheck:!1,autoComplete:"off"})]}),a.jsxs("label",{children:[m?"Connection":E==="every_application"?"Default Connection":"Bind to Connection",a.jsx("select",{value:b,onChange:W=>C(W.target.value),children:v.map(W=>a.jsx("option",{children:W},W))})]}),m&&a.jsxs("label",{children:["Model to use",a.jsx("input",{value:x,onChange:W=>y(W.target.value),required:!0,maxLength:256,placeholder:"model name on that connection",spellCheck:!1,autoComplete:"off"})]}),a.jsxs("label",{children:["Note ",h?"for "+o.application:"(optional)",a.jsx("input",{value:N,onChange:W=>D(W.target.value),maxLength:2e3})]}),a.jsxs("div",{className:"actions",children:[a.jsx("button",{className:"primary",disabled:!!H||!J,children:H==="accept"?"Approving…":"Approve"}),a.jsx("button",{type:"button",className:"secondary",disabled:!!H,onClick:()=>$(Z),children:H===Z?h?"Rejecting…":"Dismissing…":h?"Reject":"Dismiss"})]})]}),o.informational&&a.jsxs("div",{className:"actions",children:[o.opens==="applications"&&r&&a.jsxs("button",{type:"button",className:"primary",onClick:r,children:["Open Integrate to pair it ",a.jsx(dn,{size:14,"aria-hidden":!0})]}),a.jsx("button",{type:"button",className:"secondary",disabled:!!H,onClick:()=>$("dismiss"),children:H==="dismiss"?"Dismissing…":"Dismiss"})]})]})}function Hv({item:o,targets:c}){const{target:u,model:r}=o.proposal;if(o.source!=="request"||!u&&!r)return null;const h=!!u&&!!c&&!c.includes(u);return a.jsxs("p",{className:"decision-ask",children:[a.jsxs("span",{children:[o.application," suggests"]})," ",r||"a model",u?" on "+u:"",h?" (no such Connection here; choose one below)":"",". You decide."]})}function $w({item:o}){const c=o.source==="request",u=o.decision,r=o.status==="accepted"?"Accepted":o.status==="dismissed"?"Dismissed":"Rejected";return a.jsxs("div",{className:"decision decided",children:[a.jsxs("div",{className:"decision-head",children:[a.jsx("h3",{className:"need",children:o.need}),a.jsxs("div",{className:"decision-meta",children:[a.jsx("span",{className:"badge"+(o.status==="accepted"?" good":""),children:r}),a.jsx("span",{className:"badge",children:c?"Request from "+o.application:"Suggestion"}),a.jsxs("span",{className:"badge",children:["For ",Wi(u?.scope||o.scope,o.application)]})]})]}),c&&o.details&&a.jsxs("p",{className:"decision-ask",children:[a.jsx("span",{children:"Asked for"})," ",Uv(o.details)]}),a.jsx(Hv,{item:o}),u&&a.jsxs("p",{className:"evidence",children:[r," by ",u.by," on ",Rt(u.at_ms),u.note?": “"+u.note+"”":".",o.status==="accepted"?" The Change is in Applied changes below, where you can revert it.":""]}),c&&a.jsxs("details",{className:"decision-more",children:[a.jsxs("summary",{children:[a.jsx(un,{size:14,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),"Details and reasons"]}),a.jsxs("dl",{className:"request-text",children:[o.details&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Details"}),a.jsx("dd",{children:o.details})]}),o.why&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Why"}),a.jsx("dd",{children:o.why})]}),o.scope_case&&a.jsxs(a.Fragment,{children:[a.jsxs("dt",{children:["Why ",Wi(o.scope,o.application)]}),a.jsx("dd",{children:o.scope_case})]}),o.filed_by&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Filed by (as stated)"}),a.jsx("dd",{children:o.filed_by})]})]})]})]})}function eb({tick:o,onChange:c,targets:u=[],onOpenApplications:r}){const[h,m]=f.useState([]),[g,v]=f.useState([]),[b,C]=f.useState([]),[x,y]=f.useState(null),[w,A]=f.useState(!1),[E,_]=f.useState(""),[N,D]=f.useState(""),[H,G]=f.useState(""),[Q,Y]=f.useState(0),[S,U]=f.useState(null),[Z,$]=f.useState([]),[oe,be]=f.useState([]),[re,J]=f.useState("pending"),[ve,W]=f.useState(null),[F,ne]=f.useState(!0),[le,je]=f.useState(null);f.useEffect(()=>{Vt("GET","upstream").then(U).catch(()=>U(null)),Vt("GET","changes").then(I=>{y(I.changes||[]),A(!!I.revertable)}).catch(I=>D(I.message)),Vt("GET","decisions").then(I=>{m((I.items||[]).filter(ee=>!ee.issue)),v(I.demand||[])}).catch(()=>{m([]),v([])});const j=I=>(I.items||[]).filter(ee=>!ee.issue);Vt("GET","decisions?status=accepted").then(I=>$(j(I))).catch(()=>$([])),Vt("GET","decisions?status=rejected").then(I=>be(j(I))).catch(()=>be([])),Vt("GET","issues").then(I=>{C(I.issues||[]),W(!!I.auto_fix),ne(I.adapting!==!1),je(zv(I))}).catch(()=>{C([]),W(null)})},[o,Q]);const Ae=async j=>{_(j.id),D(""),G("");try{await Vt("POST","changes/"+encodeURIComponent(j.id)+"/revert"),G("Reverted “"+j.what+"”. The previous configuration is live."),Y(I=>I+1),c()}catch(I){D(I.message)}finally{_("")}},Qe=async j=>{_("auto-fix"),D(""),G("");try{await Vt("PUT","adapting/auto-fix",{enabled:j}),G(j?"Automatic fixes are on: aiport applies an Issue’s aiport fix itself when only one Connection serves the model. You can revert each one below.":"Automatic fixes are off: every aiport fix waits for your decision."),Y(I=>I+1),c()}catch(I){D(I.message)}finally{_("")}};f.useEffect(()=>{const j=window.location.hash.slice(1);if(!j.startsWith("issue-")||!b.length)return;const I=document.getElementById(j);I&&(I.scrollIntoView({behavior:"smooth",block:"start"}),I.focus({preventScroll:!0}),history.replaceState(null,"",window.location.pathname+window.location.search))},[b]);const Ze=new Map((x||[]).map(j=>[j.id,j])),dt=j=>{G(j),D(""),Y(I=>I+1),c()};return a.jsxs(a.Fragment,{children:[!!b.length&&a.jsxs("section",{className:"panel changes decisions","aria-label":"Issues",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:[a.jsx(wx,{size:16,"aria-hidden":!0})," Issues",a.jsx(wt,{section:"Issues"})]}),a.jsx("span",{children:"Problems aiport keeps seeing for one Application · they close themselves when its calls succeed again"})]}),b.map(j=>a.jsx(Ww,{issue:j,targets:u,explain:le,onDecided:dt,onExplained:()=>Y(I=>I+1)},j.id+j.state+(j.fix?.suggestion||"")))]}),(F||h.length+Z.length+oe.length>0)&&a.jsxs("section",{className:"panel changes suggestions decisions","aria-label":"Suggestions and Change requests",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:[a.jsx(Ex,{size:16,"aria-hidden":!0})," Suggestions and Change requests",a.jsx(wt,{section:"What waits for your decision"})]}),a.jsx("div",{className:"task-filters decision-filters",role:"tablist","aria-label":"Show",children:[["pending","Waiting",h.length],["accepted","Accepted",Z.length],["rejected","Rejected",oe.length]].map(([j,I,ee])=>a.jsxs("button",{type:"button",role:"tab","aria-selected":re===j,className:re===j?"active":"",onClick:()=>J(j),children:[I,a.jsx("span",{children:ee})]},j))})]}),re==="pending"&&(h.length?h.map(j=>a.jsx(Pw,{item:j,targets:u,onDecided:dt,onOpenApplications:r},j.id)):a.jsx("p",{className:"quiet-empty",children:"Nothing waits for your decision. Suggestions from request history and Change requests from Applications appear here."})),re!=="pending"&&((re==="accepted"?Z:oe).length?(re==="accepted"?Z:oe).map(j=>a.jsx($w,{item:j},j.id)):a.jsxs("p",{className:"quiet-empty",children:["No ",re," items. Decided items are kept as long as request history."]}))]}),!!g.length&&a.jsxs("section",{className:"panel changes","aria-label":"Demand for things aiport cannot do yet",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Asked for, not supported yet"}),a.jsx("span",{children:"Change requests aiport cannot apply · kept and counted, no decision needed"})]}),a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Kind"}),a.jsx("th",{children:"Requests"}),a.jsx("th",{children:"Applications"}),a.jsx("th",{children:"Last asked"})]})}),a.jsx("tbody",{children:g.map(j=>a.jsxs("tr",{children:[a.jsx("td",{children:j.kind}),a.jsx("td",{children:j.requests}),a.jsx("td",{children:j.applications}),a.jsx("td",{children:Rt(j.last_at_ms)})]},j.kind))})]})}),S&&!!S.suggestions.length&&a.jsx("ul",{className:"askers","aria-label":"Suggest to aiport",children:S.suggestions.map(j=>a.jsxs("li",{children:[a.jsx("strong",{children:j.kind})," ",a.jsx("span",{className:"caption",children:"kind named by an Application"})," ",a.jsx(_m,{label:"Suggest to aiport",draft:j,what:{suggest:j.kind},upstream:S})]},j.kind))})]}),S&&!!S.shares.length&&a.jsxs("section",{className:"panel changes","aria-label":"Extensions",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:[a.jsx(Lx,{size:16,"aria-hidden":!0})," Extensions"]}),a.jsx("span",{children:"Listed in aiport.yaml · share one upstream for others to use"})]}),a.jsx("ul",{className:"askers",children:S.shares.map(j=>a.jsxs("li",{children:[a.jsx("strong",{children:j.extension})," (",j.hook,") ",a.jsx(_m,{label:"Share upstream",draft:j,what:{share:j.extension},upstream:S})]},j.extension))})]}),a.jsxs("section",{className:"panel changes",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:["Applied changes",a.jsx(wt,{section:"Applied changes and Revert"})]}),a.jsxs("div",{className:"header-actions",children:[a.jsx("span",{className:"caption",children:"Newest first · kept across restarts"}),a.jsx(cw,{})]})]}),ve!==null&&a.jsxs("label",{className:"check-label compact",title:"adapting.auto_fix in aiport.yaml",children:[a.jsx("input",{type:"checkbox",checked:ve,disabled:!!E||!F,onChange:j=>Qe(j.target.checked)}),a.jsxs("span",{children:["Fix Issues automatically when the fix is clear. You can revert any fix below, and a reverted fix is not applied again.",!F&&" Adapting is off, so nothing is applied."]})]}),N&&a.jsx("div",{className:"notice error",role:"alert",children:N}),H&&a.jsx("div",{className:"notice",role:"status",children:H}),x&&!x.length&&a.jsxs("div",{className:"empty",children:[a.jsx("span",{className:"empty-icon",children:a.jsx(Ig,{size:20,strokeWidth:1.75,"aria-hidden":!0})}),a.jsx("h3",{children:"No changes yet"}),a.jsx("p",{children:"Configuration edits made through aiport appear here, with the file before and after kept so you can revert them."})]}),!!x?.length&&a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Time"}),a.jsx("th",{children:"Change"}),a.jsx("th",{children:"Made by"}),a.jsx("th",{children:"Status"}),a.jsx("th",{})]})}),a.jsx("tbody",{children:x.map(j=>{const I=j.reverted_by?Ze.get(j.reverted_by):void 0;return a.jsxs("tr",{children:[a.jsx("td",{title:new Date(j.at_ms).toISOString(),children:Rt(j.at_ms)}),a.jsx("td",{children:j.by===Dv?"aiport fixed: "+j.what:j.what}),a.jsx("td",{children:j.by}),a.jsx("td",{children:j.reverted?a.jsxs("span",{className:"badge",title:I?"Reverted by "+I.by:void 0,children:["Reverted",j.reverted_at_ms?" "+Rt(j.reverted_at_ms):""]}):j.revert_of?a.jsx("span",{className:"badge",children:"Revert"}):a.jsx("span",{className:"badge good",children:"Applied"})}),a.jsx("td",{children:w&&!j.reverted&&a.jsxs("button",{className:"text-button",disabled:!!E,onClick:()=>Ae(j),"aria-label":"Revert "+j.what,children:[a.jsx(Id,{size:14,"aria-hidden":!0})," ",E===j.id?"Reverting…":"Revert"]})})]},j.id)})})]})}),a.jsx("p",{className:"caption",children:"Revert restores the configuration file exactly as it was before the change and makes it live at once. If the configuration was edited after a change, revert the newer changes first."})]})]})}function Iv(o,c){const[u,r]=f.useState(null),[h,m]=f.useState("");return f.useEffect(()=>{let g=!0;const v=o?"?application="+encodeURIComponent(o):"";return fetch("/api/v1/routing"+v,{credentials:"same-origin"}).then(b=>{if(!b.ok)throw new Error;return b.json()}).then(b=>{g&&(r(b.applications||[]),m(""))}).catch(()=>{g&&m("Could not load the Routing map.")}),()=>{g=!1}},[o,c]),{entries:u,error:h}}const tb=o=>o==="anthropic"?"Anthropic":"OpenAI",nb=o=>o.kind==="level"?"Level":o.kind==="default"?"Protocol default":o.shared?"Shared task name":"Own task name",Md=["low","medium","high"];function Gv(o){const c=o.effort||{};return Md.filter(u=>c[u]).map(u=>[u,c[u]])}function ab({application:o,r:c,onDone:u,onCancel:r}){const[h,m]=f.useState(c.effort||{}),[g,v]=f.useState(!1),[b,C]=f.useState("");async function x(w){v(!0),C("");try{const A={application:c.shared?"":o,alias:c.name,connection:c.connection,effort:Object.fromEntries(Md.map(N=>[N,(w[N]||"").trim()]).filter(([,N])=>N))},E=await fetch("/api/v1/aliases/effort",{method:"PUT",credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:JSON.stringify(A)}),_=await E.json().catch(()=>({}));if(!E.ok)throw new Error(_.error?.message||"The change was not saved ("+E.status+").");u((Object.keys(A.effort).length?"Saved the per-effort models of ":"Cleared the per-effort models of ")+c.name+" as a Change. Revert it in Changes.")}catch(A){C(A.message)}finally{v(!1)}}const y=w=>{w.preventDefault(),x(h)};return a.jsxs("form",{className:"conn-form effort-editor",onSubmit:y,"aria-label":"Per-effort models of "+c.name,children:[a.jsxs("p",{className:"muted",children:[c.shared?"Shared task name: these apply to every Application that sends ":"Own task name of "+o+": ",a.jsx("code",{children:c.name}),". Models of ",c.connection,"; an empty one uses ",a.jsx("code",{children:c.model}),"."]}),a.jsx("div",{className:"level-efforts",children:Md.map(w=>a.jsxs("label",{children:[w[0].toUpperCase()+w.slice(1),a.jsx("input",{value:h[w]||"",onChange:A=>m(E=>({...E,[w]:A.target.value})),maxLength:256,placeholder:c.model,spellCheck:!1,autoComplete:"off"})]},w))}),b&&a.jsx("p",{className:"auth-error",role:"alert",children:b}),a.jsxs("div",{className:"form-actions",children:[a.jsx("button",{className:"primary",disabled:g,children:g?"Saving…":"Save"}),Gv(c).length>0&&a.jsx("button",{type:"button",className:"secondary",disabled:g,onClick:()=>{x({})},children:"Clear"}),a.jsx("button",{type:"button",className:"text-button",onClick:r,children:"Cancel"})]})]})}function sb({r:o,targets:c,providers:u}){const r=c.find(m=>m.id===o.connection),h=r?.provider?u.find(m=>m.id===r.provider)?.name||r.provider:"";return a.jsxs("span",{className:"route-served",children:[r?.provider&&a.jsx(Ks,{provider:r.provider,location:r.location,size:16}),a.jsx("code",{children:o.connection}),h&&a.jsx("small",{className:"muted",children:h})]})}function Bv({entry:o,targets:c,providers:u,provider:r,onChange:h}){const[m,g]=f.useState(-1),[v,b]=f.useState(""),C=w=>c.find(A=>A.id===w)?.provider||"",x=o.routes.filter(w=>!r||C(w.connection)===r);if(!x.length)return a.jsx("p",{className:"muted route-none",children:o.routes.length?"Nothing on this provider.":"No name reaches a Connection: no binding, no protocol default and no task name."});const y=w=>o.models.filter(A=>A.connection===w.connection).length;return a.jsxs(a.Fragment,{children:[a.jsxs("table",{className:"route-table","aria-label":"Names "+o.name+" can send",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Name sent"}),a.jsx("th",{children:"Kind"}),a.jsx("th",{children:"Protocol"}),a.jsx("th",{children:"Connection"}),a.jsx("th",{children:"Model sent upstream"}),h&&a.jsx("th",{children:a.jsx("span",{className:"sr-only",children:"Edit"})})]})}),a.jsx("tbody",{children:x.map((w,A)=>{const E=Gv(w),_=m===A;return a.jsxs($i.Fragment,{children:[a.jsxs("tr",{className:"route-"+w.kind,children:[a.jsx("td",{children:w.kind==="default"?a.jsx("span",{className:"muted",children:"Any other model name"}):a.jsx("code",{children:w.name})}),a.jsx("td",{children:nb(w)}),a.jsx("td",{children:tb(w.protocol)}),a.jsx("td",{children:a.jsx(sb,{r:w,targets:c,providers:u})}),a.jsx("td",{children:w.kind==="default"?a.jsxs("span",{className:"muted",children:["As sent",y(w)?" ("+y(w)+" listed)":""]}):a.jsxs(a.Fragment,{children:[a.jsx(Ld,{size:13,"aria-hidden":!0,className:"muted"})," ",a.jsx("code",{children:w.model}),E.length>0&&a.jsx("small",{className:"route-effort",children:E.map(([N,D])=>N+": "+D).join(" · ")})]})}),h&&a.jsx("td",{className:"route-edit",children:w.kind==="task_name"&&a.jsx("button",{type:"button",className:"icon-button","aria-pressed":_,onClick:()=>{b(""),g(_?-1:A)},title:_?"Close":"Per-effort models of "+w.name,"aria-label":(_?"Close the per-effort models of ":"Edit the per-effort models of ")+w.name,children:_?a.jsx(wa,{size:14,"aria-hidden":!0}):a.jsx(Nd,{size:14,"aria-hidden":!0})})})]}),_&&h&&a.jsx("tr",{className:"route-editing",children:a.jsx("td",{colSpan:6,children:a.jsx(ab,{application:o.application,r:w,onCancel:()=>g(-1),onDone:N=>{g(-1),b(N),h()}})})})]},A)})})]}),v&&a.jsx("p",{className:"auth-success",role:"status",children:v})]})}function ib({targets:o,providers:c,reload:u,onChange:r}){const[h,m]=f.useState(0),{entries:g,error:v}=Iv(null,String(u)+h),b=()=>{m(_=>_+1),r()},[C,x]=f.useState(""),[y,w]=f.useState(""),A=Array.from(new Set((g||[]).flatMap(_=>_.routes.map(N=>o.find(D=>D.id===N.connection)?.provider||"")).filter(Boolean))),E=(g||[]).filter(_=>!C||_.application===C);return a.jsxs("section",{className:"panel routing-map","aria-labelledby":"routing-map-heading",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{id:"routing-map-heading",children:[a.jsx(Bg,{size:16,"aria-hidden":!0})," Routing map",a.jsx(wt,{section:"The Routing map"})]}),a.jsx("span",{children:"Each name an Application can send, and the Connection and model that serve it"})]}),v&&a.jsx("p",{className:"auth-error",role:"alert",children:v}),!g&&!v&&a.jsx("p",{className:"muted",children:"Loading…"}),g&&g.length===0&&a.jsx("p",{className:"quiet-empty",children:"No Applications yet. When one is approved, every name it can send is listed here."}),g&&g.length>0&&a.jsxs(a.Fragment,{children:[a.jsxs("div",{className:"routing-map-tools",children:[a.jsxs("label",{children:["Application ",a.jsxs("select",{value:C,onChange:_=>x(_.target.value),children:[a.jsxs("option",{value:"",children:["All (",g.length,")"]}),g.map(_=>a.jsx("option",{value:_.application,children:_.name||_.application},_.application))]})]}),A.length>1&&a.jsxs("div",{className:"task-filters",role:"tablist","aria-label":"Provider",children:[a.jsx("button",{type:"button",role:"tab","aria-selected":!y,className:y?"":"active",onClick:()=>w(""),children:"All providers"}),A.map(_=>a.jsxs("button",{type:"button",role:"tab","aria-selected":y===_,className:y===_?"active":"",onClick:()=>w(_),children:[a.jsx(Ks,{provider:_,size:14})," ",c.find(N=>N.id===_)?.name||_]},_))]})]}),E.map(_=>a.jsxs("div",{className:"routing-map-app",children:[a.jsxs("h3",{children:[_.name||_.application," ",a.jsx("small",{children:a.jsx("code",{children:_.application})}),a.jsxs("span",{className:"muted-small",children:[_.binding?"Bound to "+_.binding:"Follows the protocol defaults"," · Catalog ",_.version]})]}),a.jsx(Bv,{entry:_,targets:o,providers:c,provider:y,onChange:b})]},_.application))]})]})}function ob({application:o,targets:c,providers:u,reload:r,onChange:h}){const[m,g]=f.useState(0),{entries:v,error:b}=Iv(o,String(r)+m),[C,x]=f.useState(!1);if(b)return a.jsx("p",{className:"auth-error",role:"alert",children:b});if(!v)return a.jsx("p",{className:"muted",children:"Loading…"});const y=v[0];if(!y)return null;const w=y.changes||[],A=C?w:w.slice(0,5);return a.jsxs("div",{className:"app-catalog",children:[a.jsxs("h4",{children:["Catalog ",a.jsx("code",{children:y.version}),a.jsx(wt,{section:"The Catalog"})]}),a.jsx(Bv,{entry:y,targets:c,providers:u,onChange:()=>{g(E=>E+1),h?.()}}),y.models.length>0&&a.jsxs("details",{className:"app-catalog-models",children:[a.jsxs("summary",{children:[a.jsx(un,{size:14,"aria-hidden":!0,className:"chevron"})," ",y.models.length," models listed by ",y.connections.join(", ")]}),a.jsx("p",{children:y.models.map(E=>E.id).join(", ")})]}),a.jsx("h4",{children:"Recent Catalog changes"}),w.length===0?a.jsx("p",{className:"muted",children:"None kept. A Change that adds, removes or moves a name this Application can send is listed here."}):a.jsx("ul",{className:"app-catalog-changes",children:A.map(E=>a.jsxs("li",{children:[a.jsxs("span",{className:"muted-small",children:[new Date(E.at_ms).toLocaleString()," · Change ",E.change_id,E.revert_of?" (revert of "+E.revert_of+")":""]}),Ev(E).join("; ")||"Catalog version changed"]},E.change_id+E.at_ms))}),w.length>5&&a.jsx("button",{type:"button",className:"text-button",onClick:()=>x(E=>!E),children:C?"Show fewer":"Show all "+w.length})]})}async function lb(o,c,u){const r=await fetch("/api/v1/targets/"+encodeURIComponent(o)+"/api-key",{method:c,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:c==="PUT"?JSON.stringify({api_key:u}):void 0}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Could not update the API key.");return h}const rb={saved:["Key saved","good"],environment:["Key from environment","good"],missing:["Key missing","warn"],not_needed:["No key needed","good"]};function cb({target:o,envName:c,source:u,onChange:r}){const[h,m]=f.useState(""),[g,v]=f.useState(!1),[b,C]=f.useState(""),[x,y]=f.useState("");async function w(_){v(!0),C(""),y("");try{const N=await lb(o,_,h.trim());m(""),y(_==="PUT"?"Saved. Applies to new requests immediately.":N.secret_source[o]==="environment"?"Removed. Using "+c+" from the server environment.":"Removed."),r(N)}catch(N){C(N.message)}finally{v(!1)}}const A=_=>{_.preventDefault(),h.trim()&&w("PUT")},E="key-"+o;return a.jsxs("form",{className:"key-field",onSubmit:A,children:[a.jsx("label",{htmlFor:E,children:u==="missing"?"API key":"Replace API key"}),a.jsxs("div",{className:"key-row",children:[a.jsx("input",{id:E,type:"password",autoComplete:"off",spellCheck:!1,value:h,onChange:_=>m(_.target.value),placeholder:u==="saved"?"•••••••• saved — paste to replace":u==="environment"?"Using "+c+" — paste to override":"Paste your "+c,maxLength:4e3}),a.jsx("button",{className:"secondary",disabled:g||!h.trim(),children:g?"Saving…":"Save key"})]}),b&&a.jsx("p",{className:"auth-error",role:"alert",children:b}),x&&a.jsx("p",{className:"auth-success",role:"status",children:x}),u==="saved"&&a.jsx("button",{type:"button",className:"text-button key-remove",disabled:g,onClick:()=>{w("DELETE")},children:"Remove saved key"})]})}function Hm(o){if(!o)return;const c=o.listen.slice(o.listen.lastIndexOf(":")+1),u=o.listen.slice(0,o.listen.lastIndexOf(":"));return(!u||u==="0.0.0.0"||u==="[::]"?"127.0.0.1":u)+":"+c}function Im(o){if(!o)return;const c=o.hostnames?.[0];return c?c+":"+o.listen.slice(o.listen.lastIndexOf(":")+1):o.listen}const db={local:{title:"Local",icon:Cx,blurb:"Model servers on this machine. No API key; requests never leave your computer."},remote:{title:"Remote APIs",icon:mg,blurb:"Hosted providers. Each needs an API key; requests are sent to the provider over HTTPS."}};function ub({text:o,label:c}){const[u,r]=f.useState(!1);return a.jsxs("div",{className:"copy-block",children:[a.jsx("pre",{children:o}),a.jsx("button",{type:"button",className:"icon-button","aria-label":"Copy "+c,title:"Copy",onClick:()=>{navigator.clipboard?.writeText(o).then(()=>{r(!0),setTimeout(()=>r(!1),1200)})},children:u?a.jsx(Zs,{size:14,"aria-hidden":!0}):a.jsx(Fa,{size:14,"aria-hidden":!0})})]})}const hb=["us-east-1","us-east-2","us-west-2","ca-central-1","sa-east-1","eu-central-1","eu-west-1","eu-west-2","eu-west-3","eu-north-1","eu-south-1","ap-south-1","ap-northeast-1","ap-northeast-2","ap-southeast-1","ap-southeast-2"],xa=o=>o==="openai"?"OpenAI":"Anthropic",Yv=o=>[...new Set(o.endpoints.flatMap(c=>Object.keys(c.urls)))],pb=o=>!!o&&(o.regional||o.endpoints.length>1||Yv(o).length>1);async function kn(o,c,u){const r=await fetch("/api/v1/"+c,{method:o,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:u?JSON.stringify(u):void 0}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"The change was not saved ("+r.status+").");return h}const fb=o=>o.state==="ok"?"good":o.ok?"warn":"bad",mb=o=>o.state==="ok"?"Healthy":o.state==="no_model_list"?"Reachable":o.state==="no_key"?"Needs a key":o.state==="disabled"?"Disabled":"Unhealthy";function gb({state:o}){return o?o==="running"?a.jsx("p",{className:"check-result",role:"status",children:"Checking…"}):"error"in o?a.jsxs("p",{className:"check-result",role:"alert",children:[a.jsx("span",{className:"badge bad",children:"Check failed"})," ",o.error]}):a.jsxs("p",{className:"check-result",role:"status",children:[a.jsx("span",{className:"badge "+fb(o),children:mb(o)})," ",o.message,o.models!==void 0&&" "+o.models+" model"+(o.models===1?"":"s")+" listed."," ",a.jsxs("small",{children:[o.latency_ms," ms · ",new Date(o.checked_at_ms).toLocaleTimeString()]}),o.reason&&a.jsxs("span",{className:"check-reason",children:[o.reason.explanation," ",o.reason.fix&&a.jsxs(a.Fragment,{children:[a.jsx("b",{children:"What to do:"})," ",o.reason.fix]})]})]}):null}function vb({t:o,check:c}){const[u,r]=f.useState(null),[h,m]=f.useState(!1),[g,v]=f.useState(!1),[b,C]=f.useState("");async function x(E){v(!0),C("");try{r(await kn("GET","targets/"+encodeURIComponent(o.id)+"/models"+(E?"?refresh=1":""))),m(!0)}catch(_){C(_.message)}finally{v(!1)}}if(o.disabled)return null;const y=c&&typeof c=="object"&&!("error"in c)?c.models:void 0,w=u?.listed?u.models.length:y,A=w===void 0?"Models":w+" model"+(w===1?"":"s");return a.jsxs("div",{className:"conn-models",children:[a.jsx("button",{type:"button",className:"text-button","aria-expanded":h,disabled:g,onClick:()=>h?m(!1):x(!1),title:"Lists this connection's models with GET /models. Generates nothing and costs nothing.",children:g?"Listing models…":h?"Hide models":u||w!==void 0?"Show "+A.toLowerCase():"Show models"}),b&&a.jsx("p",{className:"auth-error",role:"alert",children:b}),h&&u&&(u.listed?a.jsxs("div",{className:"model-list",children:[a.jsxs("p",{className:"caption",children:[A," · listed ",new Date(u.listed_at_ms).toLocaleTimeString()," ",a.jsx("button",{type:"button",className:"text-button",disabled:g,onClick:()=>x(!0),children:"Refresh"})]}),u.models.length?a.jsx("ul",{children:u.models.map(E=>a.jsx("li",{children:a.jsx("code",{children:E})},E))}):a.jsx("p",{className:"caption",children:"This connection lists no models."})]}):a.jsxs("p",{className:"caption",role:"status",children:["The model list could not be read (",u.state.replaceAll("_"," "),"). aiport treats its models as unknown. Use ",a.jsx("b",{children:"Check connection"})," to see why. ",a.jsx("button",{type:"button",className:"text-button",disabled:g,onClick:()=>x(!0),children:"Try again"})]}))]})}function Xv(o,c,u){const[r,h]=f.useState(u?.region||""),[m,g]=f.useState(u?.endpoint||o.endpoints[0].id),[v,b]=f.useState(u?.protocol||o.protocol),C=o.endpoints.find(_=>_.id===m)||o.endpoints[0],x=Object.keys(C.urls),y=x.includes(v)?v:x[0],w=(C.urls[y]||"").replaceAll("{region}",r||"{region}"),A={region:r,endpoint:C.id===o.endpoints[0].id?"":C.id,protocol:y===o.protocol?"":y};return{fields:a.jsxs(a.Fragment,{children:[o.regional&&a.jsxs("label",{children:["Region",a.jsx("input",{list:c+"-regions",value:r,onChange:_=>h(_.target.value.trim()),placeholder:"us-east-1",required:!0,pattern:"[a-z]{2,4}(-[a-z0-9]+){1,3}",maxLength:32,spellCheck:!1,autoComplete:"off"}),a.jsx("datalist",{id:c+"-regions",children:hb.map(_=>a.jsx("option",{value:_},_))})]}),o.endpoints.length>1&&a.jsxs("fieldset",{children:[a.jsx("legend",{children:"Endpoint"}),a.jsx("div",{className:"segmented",role:"radiogroup",style:{gridTemplateColumns:`repeat(${o.endpoints.length},1fr)`},children:o.endpoints.map((_,N)=>a.jsxs("button",{type:"button",role:"radio","aria-checked":C.id===_.id,className:C.id===_.id?"on":"",onClick:()=>g(_.id),children:[_.name,N===0?" · default":""]},_.id))}),C.note&&a.jsx("small",{children:C.note})]}),x.length>1&&a.jsxs("label",{children:["Protocol",a.jsx("select",{value:y,onChange:_=>b(_.target.value),children:x.map(_=>a.jsxs("option",{value:_,children:[xa(_),_==="anthropic"?" (Messages API · Claude Code)":" (Chat Completions, Responses)"]},_))})]}),a.jsxs("div",{className:"conn-preview",children:[a.jsx("span",{children:"Requests go to"}),a.jsx("code",{children:w})]})]}),value:A,protocol:y,ready:!o.regional||!!r}}function eo({error:o,message:c}){return a.jsxs(a.Fragment,{children:[o&&a.jsx("p",{className:"auth-error",role:"alert",children:o}),c&&a.jsx("p",{className:"auth-success",role:"status",children:c})]})}function yb({t:o,spec:c,onSaved:u}){const r=Xv(c,"edit-"+o.id,{region:o.region,endpoint:o.endpoint,protocol:o.protocol}),[h,m]=f.useState(!1),[g,v]=f.useState(""),[b,C]=f.useState(""),x=r.value.region!==(o.region||"")||r.value.endpoint!==(o.endpoint||"")||r.protocol!==o.protocol;async function y(w){w.preventDefault(),m(!0),v(""),C("");try{await kn("PATCH","targets/"+encodeURIComponent(o.id),r.value),C("Saved. New requests use it now."),u()}catch(A){v(A.message)}finally{m(!1)}}return o.base_url?a.jsxs("p",{className:"caption conn-note",children:["This connection uses a custom ",a.jsx("code",{children:"base_url"}),"; edit it in the YAML file."]}):a.jsxs("form",{className:"conn-form",onSubmit:y,children:[r.fields,a.jsx(eo,{error:g,message:b}),a.jsx("button",{className:"secondary",disabled:h||!x||!r.ready,children:h?"Saving…":"Save connection"})]})}function Vv(o,c){if(!c.has(o))return o;for(let u=2;;u++)if(!c.has(o+"-"+u))return o+"-"+u}function xb({spec:o,taken:c,onDone:u}){const r=Xv(o,"add-"+o.id),h=Vv(o.regional&&r.value.region&&c.has(o.id)?o.id+"-"+r.value.region:o.id,c),[m,g]=f.useState(""),[v,b]=f.useState(!1),[C,x]=f.useState("");async function y(w){w.preventDefault(),b(!0),x("");try{await kn("POST","targets",{id:m||h,provider:o.id,...r.value}),u()}catch(A){x(A.message)}finally{b(!1)}}return a.jsxs("form",{className:"conn-form add-form",onSubmit:y,children:[o.note&&a.jsx("small",{children:o.note}),a.jsxs("label",{children:["Connection id ",a.jsx("small",{children:"Shown in activity and used in routing"}),a.jsx("input",{value:m,onChange:w=>g(w.target.value.trim()),placeholder:h,pattern:"[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}",maxLength:64,spellCheck:!1,autoComplete:"off"})]}),r.fields,a.jsx(eo,{error:C,message:""}),a.jsx("button",{className:"primary",disabled:v||!r.ready,children:v?"Adding…":"Add "+o.name}),o.api_key_env&&a.jsx("small",{children:"After adding, enter the API key on its card."})]})}function wb({location:o,taken:c,onDone:u}){const[r,h]=f.useState(""),[m,g]=f.useState(""),[v,b]=f.useState("openai"),[C,x]=f.useState(o==="local"?"http://127.0.0.1:8000/v1":"https://"),[y,w]=f.useState(""),[A,E]=f.useState(!1),[_,N]=f.useState("");async function D(H){H.preventDefault(),E(!0),N("");try{await kn("POST","targets",{id:r||m,provider:m,protocol:v,base_url:C,api_key_env:y}),u()}catch(G){N(G.message)}finally{E(!1)}}return a.jsxs("form",{className:"conn-form add-form",onSubmit:D,children:[a.jsx("small",{children:"Any server that speaks the OpenAI or Anthropic API natively. aiport does not translate between them."}),a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Provider name",a.jsx("input",{value:m,onChange:H=>g(H.target.value.trim()),placeholder:o==="local"?"vllm":"together",required:!0,pattern:"[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}",spellCheck:!1})]}),a.jsxs("label",{children:["Connection id",a.jsx("input",{value:r,onChange:H=>h(H.target.value.trim()),placeholder:m?Vv(m,c):"same as name",pattern:"[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}",spellCheck:!1})]})]}),a.jsxs("label",{children:["Base URL ",a.jsx("small",{children:"API prefix, usually ending in /v1. Remote URLs must use HTTPS."}),a.jsx("input",{value:C,onChange:H=>x(H.target.value.trim()),required:!0,type:"url",spellCheck:!1})]}),a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Protocol",a.jsxs("select",{value:v,onChange:H=>b(H.target.value),children:[a.jsx("option",{value:"openai",children:"OpenAI"}),a.jsx("option",{value:"anthropic",children:"Anthropic"})]})]}),a.jsxs("label",{children:["Key variable ",a.jsx("small",{children:"Optional; name only, e.g. TOGETHER_API_KEY"}),a.jsx("input",{value:y,onChange:H=>w(H.target.value.trim().toUpperCase()),pattern:"[A-Za-z_][A-Za-z0-9_]*",spellCheck:!1})]})]}),a.jsx(eo,{error:_,message:""}),a.jsx("button",{className:"primary",disabled:A,children:A?"Adding…":"Add connection"})]})}const bb=o=>{const c=o.model?" ("+o.model+")":"";switch(o.kind){case"default":return xa(o.protocol||"openai")+"-protocol default";case"binding":return o.application+" is bound to it";case"alias":return"Shared Alias "+o.alias+c;case"application_alias":return o.application+"'s Alias "+o.alias+c;case"level":return"Level "+o.alias+" ("+xa(o.protocol||"openai")+")"+c;case"application_level":return o.application+"'s Level "+o.alias+" ("+xa(o.protocol||"openai")+")"+c;case"explain":return"Explanations"+c;case"agent_runs":return"Agent runs' model calls"+c;default:return"Compression evaluator"+c}},jb={default:"None (no default)",binding:"None (follow the defaults)",alias:"None (remove the Alias)",application_alias:"None (remove the Alias)",level:"None (Level not mapped)",application_level:"None (use the shared mapping)",explain:"None (Explanations off)",agent_runs:"None (agents use their own login)",evaluation:"None (evaluator off)"},Ab=o=>[...o.removed.map(c=>c.name+" removed"),...o.moved.map(c=>c.name+" moves to "+c.to.connection),...o.routing.map(c=>(c.of==="binding"?"binding":xa(c.of)+" default")+": "+(c.after||"none"))].join(" · ");function kb({t:o,targets:c,onClose:u,onRemoved:r}){const h=f.useRef(null),[m,g]=f.useState(null),[v,b]=f.useState({}),[C,x]=f.useState([]),[y,w]=f.useState(""),[A,E]=f.useState(!1),[_,N]=f.useState(""),D="targets/"+encodeURIComponent(o.id)+"/dependents";f.useEffect(()=>{h.current?.showModal()},[]),f.useEffect(()=>{fetch("/api/v1/"+D,{credentials:"same-origin"}).then(async Y=>{const S=await Y.json().catch(()=>({}));if(!Y.ok)throw new Error(S.error?.message||"Could not list what uses this connection ("+Y.status+").");const U=S;g(U),x(U.applications),w(U.problem||""),b(Object.fromEntries(U.dependents.map(Z=>[Z.id,Z.suggestion])))}).catch(Y=>N(Y.message))},[D]);async function H(Y,S){const U={...v,[Y]:S};b(U);try{const Z=await kn("POST",D,{plan:U});x(Z.applications),w(Z.problem||"")}catch(Z){w(Z.message)}}async function G(){E(!0),N("");try{await kn("DELETE","targets/"+encodeURIComponent(o.id),{plan:v}),r()}catch(Y){N(Y.message),E(!1)}}const Q=Y=>c.filter(S=>S.id!==o.id&&!S.disabled&&(["default","agent_runs","level","application_level"].includes(Y.kind)?S.protocol===o.protocol:!0));return a.jsxs("dialog",{ref:h,className:"recovery-dialog removal-dialog",onClose:u,"aria-labelledby":"remove-"+o.id,children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{id:"remove-"+o.id,className:"with-icon",children:[a.jsx(Dl,{size:18,"aria-hidden":!0}),"Remove ",o.id]}),a.jsxs("button",{type:"button",className:"text-button",onClick:()=>h.current?.close(),"aria-label":"Close",children:["Close ",a.jsx(wa,{size:14,"aria-hidden":!0})]})]}),a.jsxs("div",{className:"recovery-content",children:[!m&&!_&&a.jsxs("p",{children:["Looking up what uses ",o.id,"…"]}),m&&(m.dependents.length===0?a.jsxs("p",{children:["Nothing uses ",o.id,". Removing it changes no routing."]}):a.jsxs(a.Fragment,{children:[a.jsxs("p",{children:["These use ",o.id,". Choose where each goes instead, or none. Each starts with the best match: a Connection that serves the same model, else the protocol default. An Alias or Level moved to another Connection keeps its model but loses its per-effort models, which name models of ",o.id,"; set them again on its new Connection."]}),a.jsx("ul",{className:"removal-list",children:m.dependents.map(Y=>a.jsxs("li",{children:[a.jsx("label",{htmlFor:"dep-"+Y.id,children:bb(Y)}),a.jsxs("select",{id:"dep-"+Y.id,value:v[Y.id]??"",disabled:A,onChange:S=>{H(Y.id,S.target.value)},children:[Q(Y).map(S=>a.jsx("option",{value:S.id,children:S.id},S.id)),a.jsx("option",{value:"",children:jb[Y.kind]||"None"})]})]},Y.id))})]})),m&&a.jsxs("div",{className:"removal-affected",children:[a.jsxs("h3",{className:"with-icon",children:[a.jsx(Qx,{size:14,"aria-hidden":!0}),"Applications affected"]}),C.length===0?a.jsx("p",{className:"caption",children:"No Application's Catalog changes."}):a.jsx("ul",{children:C.map(Y=>a.jsxs("li",{children:[a.jsx("strong",{children:Y.name||Y.application})," ",a.jsx("small",{children:Ab(Y)})]},Y.application))})]}),y&&a.jsx("p",{className:"notice-inline",role:"status",children:y}),_&&a.jsx("p",{className:"auth-error",role:"alert",children:_}),a.jsx("p",{className:"caption",children:"The re-pointing and the removal are one Change: revert it from Changes to restore everything."}),a.jsxs("div",{className:"removal-actions",children:[a.jsxs("button",{type:"button",className:"danger with-icon",disabled:A||!m,onClick:()=>{G()},children:[a.jsx(Dl,{size:14,"aria-hidden":!0}),A?"Removing…":"Remove "+o.id]}),a.jsx("button",{type:"button",className:"text-button",onClick:()=>h.current?.close(),children:"Cancel"})]})]})]})}function Cb({t:o,targets:c,defaults:u,onChange:r}){const[h,m]=f.useState(!1),[g,v]=f.useState(!1),[b,C]=f.useState(""),x=o.protocol==="openai"?"openai_target":"anthropic_target",y=u[x]===o.id;async function w(A){v(!0),C("");try{await A(),r()}catch(E){C(E.message)}finally{v(!1)}}return a.jsxs("div",{className:"card-actions",children:[!y&&!o.disabled&&a.jsxs("button",{type:"button",className:"text-button",disabled:g,onClick:()=>w(()=>kn("PUT","defaults",{...u,[x]:o.id})),children:["Make ",xa(o.protocol)," default"]}),a.jsx("button",{type:"button",className:"text-button subtle",onClick:()=>m(!0),children:"Remove"}),h&&a.jsx(kb,{t:o,targets:c,onClose:()=>m(!1),onRemoved:()=>{m(!1),r()}}),b&&a.jsx("p",{className:"auth-error",role:"alert",children:b})]})}function Tb({t:o,all:c,spec:u,src:r,role:h,defaults:m,onKeys:g,onChange:v,check:b,onCheck:C}){const[x,y]=f.useState(!1),w=u?.name||o.provider,[A,E]=o.disabled?["Disabled",""]:rb[r];return a.jsxs("article",{className:"provider"+(o.disabled?" off":""),children:[a.jsxs("header",{className:"provider-title",children:[a.jsx("span",{className:"provider-icon","aria-hidden":!0,children:a.jsx(Ks,{provider:o.provider,location:o.location,size:20})}),a.jsxs("div",{children:[a.jsx("h3",{children:w}),a.jsx("small",{children:o.id})]}),a.jsx("span",{className:"badge "+E,children:A})]}),h&&a.jsxs("span",{className:"chip",children:["Default · ",h]}),a.jsxs("dl",{className:"facts",children:[a.jsx("dt",{children:"URL"}),a.jsx("dd",{children:a.jsx("code",{children:o.url})}),a.jsx("dt",{children:"Protocol"}),a.jsx("dd",{children:xa(o.protocol)}),o.region&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Region"}),a.jsx("dd",{children:o.region})]}),u&&u.endpoints.length>1&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Endpoint"}),a.jsx("dd",{children:(u.endpoints.find(_=>_.id===o.endpoint)||u.endpoints[0]).name})]}),o.api_key_env&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Key variable"}),a.jsx("dd",{children:a.jsx("code",{children:o.api_key_env})})]}),!u&&a.jsxs(a.Fragment,{children:[a.jsx("dt",{children:"Preset"}),a.jsx("dd",{children:"Custom provider"})]})]}),pb(u)&&a.jsxs("div",{className:"conn-edit",children:[a.jsx("button",{type:"button",className:"text-button","aria-expanded":x,onClick:()=>y(_=>!_),children:x?"Hide connection settings":"Edit region, endpoint and protocol"}),x&&a.jsx(yb,{t:o,spec:u,onSaved:v})]}),o.api_key_env&&a.jsx(cb,{target:o.id,envName:o.api_key_env,source:r,onChange:g}),a.jsxs("div",{className:"conn-check",children:[a.jsxs("button",{type:"button",className:"secondary with-icon",disabled:b==="running"||o.disabled,onClick:C,title:"Sends GET /models with this connection's key. Generates nothing and costs nothing.",children:[a.jsx(Jm,{size:14,"aria-hidden":!0}),"Check connection"]}),a.jsx(gb,{state:b}),a.jsx(vb,{t:o,check:b})]}),a.jsx(Cb,{t:o,targets:c,defaults:m,onChange:v})]})}function Sb({location:o,targets:c,all:u,presets:r,providers:h,taken:m,sources:g,secrets:v,defaults:b,onKeys:C,onChange:x}){const y=db[o],w=y.icon,[A,E]=f.useState(""),[_,N]=f.useState(!1),[D,H]=f.useState({});async function G(J){H(ve=>({...ve,[J]:"running"}));try{const ve=await kn("POST","targets/"+encodeURIComponent(J)+"/check");H(W=>({...W,[J]:ve}))}catch(ve){H(W=>({...W,[J]:{error:ve.message}}))}}const Q=c.filter(J=>!J.disabled),Y=()=>{N(!0),Q.forEach(J=>G(J.id))},S=Object.values(D).filter(J=>typeof J=="object"&&("error"in J||!J.ok)).length,U=J=>g[J.id]||(J.api_key_env?v[J.id]?"environment":"missing":"not_needed"),Z=c.filter(J=>!J.disabled&&U(J)==="missing").length,$=J=>[b.openai_target===J.id&&"OpenAI protocol",b.anthropic_target===J.id&&"Anthropic protocol"].filter(Boolean).join(" · "),oe=J=>c.filter(ve=>ve.provider===J).length,be=()=>{E(""),x()},re="group-body-"+o;return a.jsxs("section",{className:"panel provider-group grad-cards"+(_?"":" collapsed"),"aria-labelledby":"group-"+o,children:[a.jsxs("div",{className:"group-heading",children:[a.jsx("span",{className:"group-icon","aria-hidden":!0,children:a.jsx(w,{size:18,strokeWidth:1.75})}),a.jsxs("div",{children:[a.jsxs("h2",{id:"group-"+o,children:[a.jsxs("button",{type:"button",className:"group-toggle","aria-expanded":_,"aria-controls":re,onClick:()=>N(J=>!J),children:[a.jsx(un,{size:16,strokeWidth:2,"aria-hidden":!0,className:"chevron"}),y.title," ",a.jsx("span",{className:"count",children:c.length})]}),a.jsx(wt,{section:"Connections and keys"})]}),a.jsx("p",{children:y.blurb})]}),a.jsxs("div",{className:"group-actions",children:[Z>0&&a.jsxs("span",{className:"badge warn",children:[Z," need",Z===1?"s":""," a key"]}),S>0&&a.jsxs("span",{className:"badge bad",children:[S," unhealthy"]}),Q.length>0&&a.jsx("button",{type:"button",className:"text-button",onClick:Y,disabled:Object.values(D).includes("running"),children:"Check all"})]})]}),_&&a.jsxs("div",{id:re,children:[c.length?a.jsx("div",{className:"provider-grid",children:c.map(J=>a.jsx("div",{className:"provider-frame",children:a.jsx(Tb,{t:J,all:u,spec:h.find(ve=>ve.id===J.provider),src:U(J),role:$(J),defaults:b,onKeys:C,onChange:x,check:D[J.id],onCheck:()=>G(J.id)})},J.id))}):a.jsxs("p",{className:"group-empty",children:["No ",o==="local"?"local model servers":"remote providers"," yet. Add one below."]}),a.jsxs("div",{className:"available",children:[a.jsx("h3",{children:"Add a connection"}),a.jsxs("ul",{children:[r.map(J=>a.jsxs("li",{children:[a.jsxs("button",{type:"button",className:"add-row"+(A===J.id?" open":""),"aria-expanded":A===J.id,onClick:()=>E(ve=>ve===J.id?"":J.id),children:[a.jsx(ql,{size:14,"aria-hidden":!0}),a.jsx(Ks,{provider:J.id,location:o,size:16}),a.jsx("strong",{children:J.name}),oe(J.id)>0&&a.jsxs("em",{className:"count",children:[oe(J.id)," added"]}),a.jsx("span",{children:[J.regional&&"Region",J.endpoints.length>1&&J.endpoints.length+" endpoints",J.api_key_env?"API key":"No key",Yv(J).map(xa).join(" or ")+" protocol"].filter(Boolean).join(" · ")})]}),A===J.id&&a.jsx(xb,{spec:J,taken:m,onDone:be})]},J.id)),a.jsxs("li",{children:[a.jsxs("button",{type:"button",className:"add-row"+(A==="custom"?" open":""),"aria-expanded":A==="custom",onClick:()=>E(J=>J==="custom"?"":"custom"),children:[a.jsx(ql,{size:14,"aria-hidden":!0}),a.jsxs("strong",{children:["Custom ",o==="local"?"local server":"provider"]}),a.jsx("span",{children:"Your own URL · OpenAI or Anthropic protocol"})]}),A==="custom"&&a.jsx(wb,{location:o,taken:m,onDone:be})]})]})]})]})]})}function Nb({config:o,providers:c,onChange:u}){const[r,h]=f.useState(o?.defaults.openai_target||""),[m,g]=f.useState(o?.defaults.anthropic_target||""),[v,b]=f.useState(!1),[C,x]=f.useState(""),[y,w]=f.useState(""),A=o?.targets||[],E=H=>(c.find(G=>G.id===H.provider)?.name||H.provider)+" · "+H.id,_=r!==(o?.defaults.openai_target||"")||m!==(o?.defaults.anthropic_target||"");async function N(H){H.preventDefault(),b(!0),x(""),w("");try{await kn("PUT","defaults",{openai_target:r,anthropic_target:m}),w("Saved. New requests use it now."),u()}catch(G){x(G.message)}finally{b(!1)}}const D=(H,G,Q)=>a.jsxs("select",{value:G,onChange:Y=>Q(Y.target.value),children:[a.jsx("option",{value:"",children:"None (requests are rejected)"}),A.filter(Y=>Y.protocol===H).map(Y=>a.jsx("option",{value:Y.id,children:E(Y)},Y.id))]});return a.jsxs("form",{className:"conn-form defaults-form",onSubmit:N,children:[a.jsxs("label",{children:["OpenAI protocol ",a.jsx("small",{children:"/v1/chat/completions, /v1/responses"}),D("openai",r,h)]}),a.jsxs("label",{children:["Anthropic protocol ",a.jsx("small",{children:"/v1/messages · Claude Code"}),D("anthropic",m,g)]}),a.jsx(eo,{error:C,message:y}),a.jsx("button",{className:"primary",disabled:v||!_,children:v?"Saving…":"Save defaults"})]})}function Vd(o,c){const u=o?.targets||[];return{targets:u,name:h=>{const m=u.find(g=>g.id===h);return m?(c.find(g=>g.id===m.provider)?.name||m.provider)+" ("+m.id+")":"Not configured"}}}function _b({config:o,providers:c,secrets:u,sources:r,onKeys:h,onChange:m}){const{targets:g}=Vd(o,c),v=new Set(g.map(C=>C.id)),b=C=>a.jsx(Sb,{location:C,targets:g.filter(x=>x.location===C),all:g,presets:c.filter(x=>x.location===C),providers:c,taken:v,sources:r,secrets:u,defaults:o?.defaults||{},onKeys:h,onChange:m});return a.jsxs(a.Fragment,{children:[b("remote"),b("local"),a.jsxs("p",{className:"caption settings-note",children:["Keys entered here apply immediately and are stored only in the private credential file next to your configuration (",a.jsx("code",{children:"aiport.yaml.env"}),"), never in YAML and never shown again. A saved key has not been verified with its provider. Adding, editing and removing connections updates your configuration file in place and applies to new requests immediately, with no restart."]})]})}function Eb({config:o,providers:c,onChange:u,queueDepth:r,onIntegrate:h}){const{targets:m}=Vd(o,c);return a.jsxs(a.Fragment,{children:[a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:["Protocol defaults",a.jsx(wt,{section:"Routing and protocol defaults"})]}),a.jsx("span",{children:"Where requests go when an app doesn't pick a connection"})]}),a.jsx(Nb,{config:o,providers:c,onChange:u},JSON.stringify(o?.defaults)+m.length)]}),a.jsx(_w,{refresh:o,defaults:o?.defaults,onChange:u}),a.jsx(ib,{targets:m,providers:c,reload:JSON.stringify(o),onChange:u}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Connect an application"}),a.jsx("span",{children:"Point any OpenAI or Anthropic SDK here"})]}),a.jsxs("dl",{children:[a.jsx("dt",{children:"OpenAI-protocol base URL"}),a.jsx("dd",{children:a.jsxs("code",{children:["http://",Hm(o),"/v1"]})}),a.jsx("dt",{children:"Anthropic-protocol base URL"}),a.jsx("dd",{children:a.jsxs("code",{children:["http://",Hm(o)]})}),a.jsx("dt",{children:"Protocol description"}),a.jsx("dd",{children:a.jsx("a",{href:"/_connect?format=markdown",target:"_blank",rel:"noreferrer",children:a.jsx("code",{children:"/_connect"})})}),a.jsx("dt",{children:"Telemetry queue"}),a.jsxs("dd",{children:[r??0," pending"]})]})]}),a.jsx(jw,{origins:o?.origins||[],targets:m,onChange:u}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Instructions for coding agents"}),h&&a.jsx("button",{type:"button",className:"secondary",onClick:h,children:"Open Integrate"})]}),a.jsx("p",{className:"muted handover-intro integrate-pointer",children:"The Application brief and the coding-agent brief, with SDK snippets and a live checklist per Application, are on the Integrate page."})]}),a.jsxs("section",{className:"panel planned",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Model allowlists"}),a.jsx("span",{className:"badge",children:"Planned"})]}),a.jsxs("p",{className:"muted",children:["Restricting which models an application may request is not built yet. What each application's names reach is built: Levels and their per-effort models are edited in the Levels panel above (an application's own mapping under Levels on its row on Integrate), and the Routing map shows where every name goes, with an edit button on each task name's row for its per-effort models (also in an application's Catalog on Integrate). Per-application task names themselves come from approved change requests or the YAML file (",a.jsx("code",{children:"origins[].aliases"}),")."]})]})]})}function Ob({config:o,providers:c}){const{name:u}=Vd(o,c);return a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Coding-agent connections"}),a.jsxs("span",{children:[o?.origins.length||0," configured"]})]}),o?.origins.length?o.origins.map(r=>a.jsxs("div",{className:"agent",children:[a.jsx("strong",{children:r.name}),a.jsx("span",{children:r.target?u(r.target):"Protocol default"}),a.jsxs("code",{children:["aiport agent ",r.id," --model YOUR_MODEL"]})]},r.id)):a.jsxs("div",{className:"quiet-empty",children:["No coding agents connected.",a.jsx("br",{}),a.jsxs("small",{children:["Create a setup with ",a.jsx("code",{children:"aiport init --agent codex,claude"}),"."]})]})]})}function zb({config:o,providers:c,onChange:u}){const r=o?.adapting?.explain||{},[h,m]=f.useState(r.connection||""),[g,v]=f.useState(r.model||""),[b,C]=f.useState(!1),[x,y]=f.useState(""),[w,A]=f.useState(""),E=(o?.targets||[]).filter(G=>!G.disabled||G.id===r.connection),_=E.find(G=>G.id===h),N=G=>c.find(Q=>Q.id===G)?.name||G,D=h!==(r.connection||"")||g!==(r.model||"");async function H(G){G.preventDefault(),C(!0),y(""),A("");try{await kn("PUT","adapting/explainer",{connection:h,model:h?g.trim():""}),A(h?"Saved. Explain on an Issue now asks "+h+".":"Explain is off until you choose a Connection."),u()}catch(Q){y(Q.message)}finally{C(!1)}}return a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("h2",{children:["Explanations",a.jsx(wt,{section:"Explain"})]}),a.jsx("span",{children:"The model that explains an Issue when no Coding agent does"})]}),a.jsxs("form",{className:"conn-form settings-body",onSubmit:H,children:[a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Connection",a.jsxs("select",{value:h,onChange:G=>m(G.target.value),children:[a.jsx("option",{value:"",children:"None (Explain is unavailable)"}),E.map(G=>a.jsxs("option",{value:G.id,children:[N(G.provider)," · ",G.id]},G.id))]})]}),a.jsxs("label",{children:["Model",a.jsx("input",{value:g,onChange:G=>v(G.target.value),disabled:!h,required:!!h,maxLength:256,placeholder:_?.location==="local"?"llama3:8b":"model id",spellCheck:!1,autoComplete:"off"})]})]}),a.jsxs("small",{children:["Any model this Connection serves. Explain sends only the Issue’s evidence: error codes, counts, the time window, request ids, the model name, Connection ids and aiport’s diagnosis. Never keys, tokens or anything an Application wrote. The call shows in Activity under the Application ",a.jsx("code",{children:"aiport"}),"."]}),_?.location==="remote"&&a.jsxs("p",{className:"conn-note notice-inline",children:["Explanations are billed to the ",N(_.provider)," key. Choose a local Connection to keep them on this machine."]}),o?.adapting?.disabled&&a.jsx("p",{className:"conn-note notice-inline",children:"Adapting is turned off in the configuration, so Explain is unavailable."}),a.jsx(eo,{error:x,message:w}),a.jsx("div",{className:"form-actions",children:a.jsx("button",{className:"primary",disabled:b||!D||!!h&&!g.trim(),children:b?"Saving…":"Save"})})]})]})}function Rb({config:o,providers:c,onChange:u}){return a.jsxs(a.Fragment,{children:[a.jsx(Jx,{username:o?.auth.username||"",onChange:u}),a.jsx(Cw,{tick:0,targets:o?.targets||[],onChange:u}),a.jsx(zb,{config:o,providers:c,onChange:u},JSON.stringify(o?.adapting?.explain||{})),a.jsxs("section",{className:"panel setup",children:[a.jsx("p",{className:"eyebrow",children:"CONFIGURATION"}),a.jsx("h2",{children:"One configuration file. Clear ownership."}),a.jsxs("p",{className:"muted",children:["Providers, routing, Levels and their per-effort models, the per-effort models of task names (on the Routing map) and each application's coding agent are managed from their pages; task names (aliases) themselves arrive as approved Changes or are edited, with the listen address, in your YAML file. Validate hand edits with ",a.jsx("code",{children:"aiport check"}),", then restart. A new setup can be created with:"]}),a.jsx(ub,{label:"setup commands",text:`aiport init --config aiport-new.yaml
aiport check --config aiport-new.yaml
aiport serve --config aiport-new.yaml`})]})]})}const Za={pending:"Pending",in_progress:"In progress",blocked:"Blocked",completed:"Completed",failed:"Failed",cancelled:"Cancelled"},Mb={pending:"",in_progress:"info",blocked:"warn",completed:"good",failed:"bad",cancelled:""},Qd=["None","Low","Medium","High"],Qv=o=>new Date(o).toLocaleString(void 0,{dateStyle:"medium",timeStyle:"short"});async function Zd(o,c="GET",u){const r=await fetch("/api/v1/tasks"+o,{method:c,credentials:"same-origin",headers:c==="GET"?{}:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:u?JSON.stringify(u):void 0}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Task request failed ("+r.status+").");return h}function qb({status:o}){return a.jsx("span",{className:"badge "+Mb[o],children:Za[o]})}function Zv({name:o}){const[c,u]=f.useState(!1);return a.jsx("button",{type:"button",className:"task-name",title:"Copy task name",onClick:r=>{r.stopPropagation(),navigator.clipboard?.writeText(o).then(()=>{u(!0),setTimeout(()=>u(!1),1200)})},children:c?a.jsxs(a.Fragment,{children:[a.jsx(Zs,{size:12,"aria-hidden":!0})," Copied"]}):a.jsxs(a.Fragment,{children:[a.jsx(Fa,{size:12,"aria-hidden":!0})," ",o]})})}function Db({onCreated:o}){const[c,u]=f.useState(""),[r,h]=f.useState(""),[m,g]=f.useState(""),[v,b]=f.useState(0),[C,x]=f.useState(!1),[y,w]=f.useState(""),[A,E]=f.useState(!1);async function _(N){N.preventDefault(),x(!0),w("");try{const D=await Zd("","POST",{title:c,instructions:r,acceptance:m,priority:v});u(""),h(""),g(""),b(0),E(!1),o(D)}catch(D){w(D.message)}finally{x(!1)}}return a.jsx("section",{className:"panel task-new",children:a.jsxs("form",{onSubmit:_,children:[a.jsxs("div",{className:"task-quick",children:[a.jsx("input",{"aria-label":"New task title",placeholder:"Capture a task… e.g. Add keyboard shortcuts",value:c,onChange:N=>u(N.target.value),onFocus:()=>E(!0),maxLength:200,required:!0}),a.jsx("button",{className:"primary",disabled:C||!c.trim(),children:C?"Adding…":"Add task"})]}),A&&a.jsxs("div",{className:"task-more",children:[a.jsxs("label",{children:["Instructions",a.jsx("textarea",{rows:4,value:r,onChange:N=>h(N.target.value),placeholder:"What needs to change, and any constraints",maxLength:2e4})]}),a.jsxs("label",{children:["Acceptance criteria",a.jsx("textarea",{rows:3,value:m,onChange:N=>g(N.target.value),placeholder:"How completion is verified",maxLength:1e4})]}),a.jsxs("label",{className:"task-priority",children:["Priority",a.jsx("select",{value:v,onChange:N=>b(Number(N.target.value)),children:Qd.map((N,D)=>a.jsx("option",{value:D,children:N},N))})]})]}),y&&a.jsx("p",{className:"auth-error",role:"alert",children:y})]})})}function Lb({task:o,onChange:c,onClose:u}){const[r,h]=f.useState(o.title),[m,g]=f.useState(o.instructions),[v,b]=f.useState(o.acceptance),[C,x]=f.useState(o.priority),[y,w]=f.useState(o.status),[A,E]=f.useState(""),[_,N]=f.useState(!1),[D,H]=f.useState(""),[G,Q]=f.useState("");f.useEffect(()=>{h(o.title),g(o.instructions),b(o.acceptance),x(o.priority),w(o.status),E("")},[o.uuid,o.updated_at_ms]);async function Y(S){S.preventDefault(),N(!0),H(""),Q("");try{const U=await Zd("/"+encodeURIComponent(o.name),"PATCH",{title:r,instructions:m,acceptance:v,priority:C,status:y,...A.trim()?{note:A}:{}});Q("Saved."),c(U)}catch(U){H(U.message)}finally{N(!1)}}return a.jsxs("section",{className:"panel task-detail","aria-label":"Task "+o.name,children:[a.jsxs("div",{className:"panel-heading",children:[a.jsxs("div",{children:[a.jsx(Zv,{name:o.name}),a.jsx("h2",{children:o.title})]}),a.jsxs("button",{className:"text-button",onClick:u,children:["Close ",a.jsx(wa,{size:14,"aria-hidden":!0})]})]}),a.jsxs("form",{onSubmit:Y,className:"task-form",children:[a.jsxs("label",{children:["Title",a.jsx("input",{value:r,onChange:S=>h(S.target.value),maxLength:200,required:!0})]}),a.jsxs("div",{className:"task-row",children:[a.jsxs("label",{children:["Status",a.jsx("select",{value:y,onChange:S=>w(S.target.value),children:Object.keys(Za).map(S=>a.jsx("option",{value:S,children:Za[S]},S))})]}),a.jsxs("label",{children:["Priority",a.jsx("select",{value:C,onChange:S=>x(Number(S.target.value)),children:Qd.map((S,U)=>a.jsx("option",{value:U,children:S},S))})]})]}),a.jsxs("label",{children:["Instructions",a.jsx("textarea",{rows:6,value:m,onChange:S=>g(S.target.value),maxLength:2e4})]}),a.jsxs("label",{children:["Acceptance criteria",a.jsx("textarea",{rows:4,value:v,onChange:S=>b(S.target.value),maxLength:1e4})]}),a.jsxs("label",{children:["Note for history (optional)",a.jsx("input",{value:A,onChange:S=>E(S.target.value),maxLength:2e3,placeholder:"e.g. blocked on API design"})]}),D&&a.jsx("p",{className:"auth-error",role:"alert",children:D}),G&&a.jsx("p",{className:"auth-success",role:"status",children:G}),a.jsx("button",{className:"primary",disabled:_,children:_?"Saving…":"Save changes"})]}),a.jsxs("div",{className:"task-history",children:[a.jsx("h3",{children:"History"}),a.jsx("ol",{children:[...o.history].reverse().map((S,U)=>a.jsxs("li",{children:[a.jsx("span",{className:"task-when",children:Qv(S.at_ms)}),a.jsxs("span",{children:[a.jsx("strong",{children:S.actor})," ",S.action==="created"?"created the task":S.action==="status"?a.jsxs(a.Fragment,{children:["changed status ",Za[S.from]||S.from," → ",Za[S.to]||S.to]}):S.action==="edited"?"edited "+(S.note||""):"added a note",S.action!=="edited"&&S.note&&a.jsxs("em",{children:[" · ",S.note]})]})]},U))})]})]})}function Ub({refresh:o}){const[c,u]=f.useState([]),[r,h]=f.useState("open"),[m,g]=f.useState(""),[v,b]=f.useState(""),[C,x]=f.useState(!0);f.useEffect(()=>{x(!0),Zd("").then(N=>{u(N.tasks),b("")}).catch(N=>b(N.message)).finally(()=>x(!1))},[o]);const y=N=>u(D=>[N,...D.filter(H=>H.uuid!==N.uuid)].sort((H,G)=>G.created_at_ms-H.created_at_ms)),w=N=>N==="pending"||N==="in_progress"||N==="blocked",A=c.filter(N=>r==="all"||(r==="open"?w(N.status):N.status===r)),E=Object.keys(Za).map(N=>[N,c.filter(D=>D.status===N).length]),_=c.find(N=>N.uuid===m);return a.jsxs(a.Fragment,{children:[v&&a.jsx("div",{className:"notice error",role:"alert",children:v}),a.jsx(Db,{onCreated:N=>{y(N),g(N.uuid)}}),a.jsx("div",{className:"task-filters",role:"tablist","aria-label":"Filter tasks",children:[["open","Open",c.filter(N=>w(N.status)).length],["all","All",c.length],...E.map(([N,D])=>[N,Za[N],D])].map(([N,D,H])=>a.jsxs("button",{role:"tab","aria-selected":r===N,className:r===N?"active":"",onClick:()=>h(N),children:[D,a.jsx("span",{children:H})]},N))}),a.jsxs("div",{className:_?"task-layout with-detail":"task-layout",children:[a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Task queue"}),a.jsxs("span",{children:[A.length," shown · you own intent and priority"]})]}),C?a.jsx("div",{className:"quiet-empty",children:"Loading tasks…"}):A.length?a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Task"}),a.jsx("th",{children:"Status"}),a.jsx("th",{children:"Priority"}),a.jsx("th",{className:"numeric",children:"Updated"})]})}),a.jsx("tbody",{children:A.map(N=>a.jsxs("tr",{className:"task-row-item"+(N.uuid===m?" selected":""),onClick:()=>g(N.uuid),tabIndex:0,onKeyDown:D=>{D.key==="Enter"&&g(N.uuid)},children:[a.jsxs("td",{children:[a.jsx("span",{className:"task-title",children:N.title}),a.jsx(Zv,{name:N.name})]}),a.jsx("td",{children:a.jsx(qb,{status:N.status})}),a.jsx("td",{children:Qd[N.priority]}),a.jsx("td",{className:"numeric",children:a.jsx("small",{children:Qv(N.updated_at_ms)})})]},N.uuid))})]})}):a.jsxs("div",{className:"empty",children:[a.jsx("span",{className:"empty-icon",children:a.jsx(Ng,{size:20,strokeWidth:1.75,"aria-hidden":!0})}),a.jsx("h3",{children:c.length?"Nothing here":"No tasks yet"}),a.jsx("p",{children:c.length?"No tasks match this filter.":"Capture work as it comes up. Each task gets a memorable name like task-amber-fox that you can hand to a coding agent."})]})]}),_&&a.jsx(Lb,{task:_,onChange:y,onClose:()=>g("")})]})]})}const Bs=(o,c=0)=>o==null?"—":(o*100).toFixed(c)+"%",ga=o=>o==null?"—":new Intl.NumberFormat().format(Math.round(o)),Hb={insufficient_data:["Not enough data",""],possible_harm:["Possible harm","bad"],no_harm_seen:["No harm seen","good"]};function Ib({trend:o,metric:c,bucketMs:u}){const r=b=>c==="feedback"?b.feedback_avg:c==="errors"?b.requests?b.error_rate:null:b.saved_ratio,h=c==="feedback"?"Average feedback score":c==="errors"?"Error rate":"Request size saved",m=b=>b==null?"no data":c==="feedback"?b.toFixed(2):Bs(b,1),g=Math.max(1e-4,...o.flatMap(b=>[r(b.compressed),c==="saved"?null:r(b.baseline)]).filter(b=>b!=null)),v=b=>new Date(b).toLocaleString(void 0,u>=864e5?{month:"short",day:"numeric"}:{hour:"2-digit",minute:"2-digit"});return o.length?a.jsxs(a.Fragment,{children:[a.jsxs("div",{className:"legend","aria-hidden":!0,children:[a.jsxs("span",{children:[a.jsx("i",{className:"swatch compressed"}),"Compressed"]}),c!=="saved"&&a.jsxs("span",{children:[a.jsx("i",{className:"swatch baseline"}),"Uncompressed"]})]}),a.jsx("div",{className:"pair-bars",role:"img","aria-label":h+" per period, compressed versus uncompressed",children:o.map(b=>{const C=r(b.compressed),x=r(b.baseline);return a.jsxs("div",{className:"pair",title:`${v(b.start_ms)}
Compressed: ${m(C)} (${b.compressed.requests} requests)`+(c==="saved"?"":`
Uncompressed: ${m(x)} (${b.baseline.requests} requests)`),children:[a.jsxs("div",{className:"pair-cols",children:[a.jsx("i",{className:"compressed",style:{height:C==null?0:Math.max(2,C/g*100)+"%"}}),c!=="saved"&&a.jsx("i",{className:"baseline",style:{height:x==null?0:Math.max(2,x/g*100)+"%"}})]}),a.jsx("small",{children:v(b.start_ms)})]},b.start_ms)})})]}):a.jsx("div",{className:"quiet-empty",children:"No inference requests in this period."})}async function Gm(o,c,u){const r=await fetch("/api/v1/"+c,{method:o,credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:u?JSON.stringify(u):void 0}),h=await r.json().catch(()=>({}));if(!r.ok)throw new Error(h.error?.message||"Request failed ("+r.status+").");return h}function Gb({config:o,providers:c,onChange:u}){const r=o?.evaluation||{},[h,m]=f.useState(r.connection||""),[g,v]=f.useState(r.model||""),[b,C]=f.useState(""),[x,y]=f.useState(""),[w,A]=f.useState(""),[E,_]=f.useState(null);f.useEffect(()=>{m(r.connection||""),v(r.model||"")},[r.connection,r.model]);const N=o?.targets||[],D=N.find(U=>U.id===h),H=U=>c.find(Z=>Z.id===U)?.name||U,G=h!==(r.connection||"")||g!==(r.model||"");async function Q(U){U.preventDefault(),C("save"),y(""),A(""),_(null);try{await Gm("PUT","evaluation",{connection:h,model:h?g:""}),A(h?"Saved. Test it to confirm the model answers.":"Evaluation turned off."),u()}catch(Z){y(Z.message)}finally{C("")}}async function Y(){C("test"),y(""),A(""),_(null);try{_(await Gm("POST","evaluation/test"))}catch(U){y(U.message)}finally{C("")}}const S=(U,Z)=>{const $=N.filter(oe=>oe.location===U&&!oe.disabled);return $.length?a.jsx("optgroup",{label:Z,children:$.map(oe=>a.jsxs("option",{value:oe.id,children:[H(oe.provider)," · ",oe.id]},oe.id))}):null};return a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Evaluator"}),a.jsx("span",{children:"The model that judges whether compression changed an answer"})]}),a.jsxs("form",{className:"conn-form defaults-form",onSubmit:Q,children:[a.jsxs("div",{className:"form-row",children:[a.jsxs("label",{children:["Connection",a.jsxs("select",{value:h,onChange:U=>m(U.target.value),children:[a.jsx("option",{value:"",children:"Off (no model-based evaluation)"}),S("local","Local"),S("remote","Remote APIs")]})]}),a.jsxs("label",{children:["Model ",a.jsx("small",{children:"Any model this connection serves"}),a.jsx("input",{value:g,onChange:U=>v(U.target.value),disabled:!h,required:!!h,maxLength:256,placeholder:D?.location==="local"?"llama3.1":"model id",spellCheck:!1})]})]}),D?.location==="remote"&&a.jsxs("p",{className:"conn-note notice-inline",children:["Evaluation sends question and answer text to ",H(D.provider)," and is billed to that key. Choose a local connection to keep it on this machine."]}),D?.location==="local"&&a.jsx("small",{children:"Evaluation stays on this machine."}),x&&a.jsx("p",{className:"auth-error",role:"alert",children:x}),w&&a.jsx("p",{className:"auth-success",role:"status",children:w}),E&&a.jsxs("div",{className:"test-result "+(E.ok?"good":"bad"),role:"status",children:[a.jsx("strong",{children:E.error?"Evaluator failed: "+E.error:E.ok?`Passed ${E.passed}/${E.total} checks in ${E.latency_ms} ms`:`Passed ${E.passed}/${E.total}; this model may be too weak to judge reliably`}),E.results?.map(U=>a.jsxs("span",{children:[U.case,": expected ",U.expected,", got ",U.got]},U.case))]}),a.jsxs("div",{className:"form-actions",children:[a.jsx("button",{className:"primary",disabled:!!b||!G||!!h&&!g.trim(),children:b==="save"?"Saving…":"Save evaluator"}),a.jsx("button",{type:"button",className:"secondary",disabled:!!b||G||!r.connection,onClick:Y,children:b==="test"?"Testing…":"Test evaluator"})]})]})]})}function Bb({tick:o,config:c,providers:u,onChange:r}){const[h,m]=f.useState("168"),[g,v]=f.useState("feedback"),[b,C]=f.useState(null),[x,y]=f.useState("");f.useEffect(()=>{const G=new AbortController;return fetch("/api/v1/compression?hours="+h,{credentials:"same-origin",signal:G.signal}).then(Q=>Q.ok?Q.json():Promise.reject(new Error("Could not load compression data ("+Q.status+")."))).then(Q=>{C(Q),y("")}).catch(Q=>{Q.name!=="AbortError"&&y(Q.message)}),()=>G.abort()},[h,o]);const w=b?.groups||[],A=b?.baseline,E=w.reduce((G,Q)=>G+Q.requests,0),_=w.reduce((G,Q)=>G+Q.bytes_before,0),N=w.reduce((G,Q)=>G+Q.bytes_after,0),D=w.reduce((G,Q)=>G+Q.est_tokens_saved,0),H=w.reduce((G,Q)=>G+Q.feedback_count,0);return a.jsxs(a.Fragment,{children:[x&&a.jsx("div",{className:"notice error",role:"alert",children:x}),b?.telemetry.state==="disabled"&&a.jsxs("div",{className:"notice",children:[a.jsx("span",{className:"dot amber"}),a.jsxs("div",{children:[a.jsx("strong",{children:"Request recording is off."}),a.jsxs("span",{children:[" Compression still works when apps ask for it, but nothing is measured. Set ",a.jsx("code",{children:"telemetry.enabled: true"})," and restart."]})]})]}),a.jsx("div",{className:"filters",children:a.jsxs("label",{children:["Period",a.jsxs("select",{value:h,onChange:G=>m(G.target.value),children:[a.jsx("option",{value:"24",children:"Last 24 hours"}),a.jsx("option",{value:"168",children:"Last 7 days"}),a.jsx("option",{value:"720",children:"Last 30 days"})]})]})}),a.jsx("section",{className:"metrics","aria-label":"Compression summary",children:[["Compressed requests",ga(E),ga(A?.requests)+" uncompressed for comparison"],["Request size saved",_?Bs(1-N/_,1):"—","Across compressed requests"],["Tokens saved",D?"≈ "+ga(D):"—","Estimated from bytes (÷4)"],["Responses scored",ga(H),E?Bs(H/E)+" of compressed requests":"Apps send feedback"]].map(([G,Q,Y])=>a.jsxs("article",{className:"metric",children:[a.jsx("span",{children:G}),a.jsx("strong",{children:Q}),a.jsx("small",{children:Y})]},G))}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Does compression hurt responses?"}),a.jsx("div",{className:"segmented metric-switch",role:"radiogroup","aria-label":"Trend metric",children:[["feedback","Feedback"],["errors","Errors"],["saved","Saved"]].map(([G,Q])=>a.jsx("button",{type:"button",role:"radio","aria-checked":g===G,className:g===G?"on":"",onClick:()=>v(G),children:Q},G))})]}),a.jsx("div",{className:"chart-body",children:a.jsx(Ib,{trend:b?.trend||[],metric:g,bucketMs:b?.bucket_ms||36e5})})]}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"By method"}),a.jsx("span",{children:"Compared with uncompressed requests in the same period"})]}),w.length?a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Methods"}),a.jsx("th",{className:"numeric",children:"Requests"}),a.jsx("th",{className:"numeric",children:"Saved"}),a.jsx("th",{className:"numeric",children:"Errors"}),a.jsx("th",{className:"numeric",children:"Latency"}),a.jsx("th",{className:"numeric",children:"Feedback"}),a.jsx("th",{children:"Verdict"})]})}),a.jsx("tbody",{children:w.map(G=>{const[Q,Y]=Hb[G.verdict||"insufficient_data"];return a.jsxs("tr",{children:[a.jsx("td",{children:a.jsx("code",{children:G.key})}),a.jsx("td",{className:"numeric",children:ga(G.requests)}),a.jsxs("td",{className:"numeric",children:[Bs(G.saved_ratio,1),a.jsxs("small",{children:["≈ ",ga(G.est_tokens_saved)," tokens"]})]}),a.jsxs("td",{className:"numeric",children:[Bs(G.error_rate,1),a.jsxs("small",{children:["vs ",Bs(A?.error_rate,1)]})]}),a.jsxs("td",{className:"numeric",children:[ga(G.avg_latency_ms)," ms",a.jsxs("small",{children:["vs ",ga(A?.avg_latency_ms)," ms"]})]}),a.jsxs("td",{className:"numeric",children:[G.feedback_avg==null?"—":G.feedback_avg.toFixed(2),a.jsxs("small",{children:[G.feedback_count," scored · vs ",A?.feedback_avg==null?"—":A.feedback_avg.toFixed(2)]})]}),a.jsxs("td",{children:[a.jsx("span",{className:"badge "+Y,title:G.verdict_reason,children:Q}),a.jsx("small",{className:"verdict-reason",children:G.verdict_reason})]})]},G.key)})})]})}):a.jsxs("div",{className:"empty",children:[a.jsx("span",{className:"empty-icon",children:a.jsx(Jg,{size:20,strokeWidth:1.75,"aria-hidden":!0})}),a.jsx("h3",{children:"No compressed requests yet"}),a.jsxs("p",{children:["Applications opt in per request with the ",a.jsx("code",{children:b?.header||"X-Aiport-Compress"})," header. Results appear here and are compared with uncompressed traffic."]})]})]}),a.jsx(Gb,{config:c,providers:u,onChange:r}),a.jsxs("div",{className:"charts",children:[a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Methods"}),a.jsx("span",{children:"Applied in this order; only message text is changed"})]}),a.jsx("ul",{className:"method-list",children:(b?.methods||[]).map(G=>a.jsxs("li",{children:[a.jsxs("div",{children:[a.jsx("strong",{children:G.name})," ",a.jsx("code",{children:G.id})]}),a.jsx("p",{children:G.description}),a.jsx("span",{className:"badge "+(G.lossless?"good":"warn"),children:G.lossless?"Meaning preserved":"Lossy"})]},G.id))})]}),a.jsxs("section",{className:"panel setup",children:[a.jsx("p",{className:"eyebrow",children:"FOR APPLICATIONS"}),a.jsx("h2",{children:"Turn it on per request"}),a.jsxs("p",{className:"muted",children:["Send the header on any chat, responses or messages call. Tool definitions and schemas are never changed. The response tells you what happened in ",a.jsx("code",{children:"X-Aiport-Compression"}),"."]}),a.jsx("pre",{children:`${b?.header||"X-Aiport-Compress"}: whitespace, json`}),a.jsxs("p",{className:"muted",children:["Then score the answer you got (0 = useless, 1 = perfect), using the ",a.jsx("code",{children:"X-Aiport-Request-Id"})," from the response. Scores power the verdict above."]}),a.jsx("pre",{children:`POST ${b?.feedback_path||"/v1/aiport/feedback"}
{"request_id": "<X-Aiport-Request-Id>", "score": 0.9}`}),a.jsx("p",{className:"caption",children:"Only sizes, outcomes and scores are recorded. Prompts and responses are never stored. Kept in memory until SQLite storage is enabled."})]})]})]})}const Ll=5,Yb=[[168,"7 days"],[720,"30 days"],[2160,"90 days"]],Xs=o=>new Intl.NumberFormat().format(o);function Kv(o){return o.up_ratio==null?"none":o.rated<Ll?"few":o.up_ratio>=.75?"good":o.up_ratio<.5?"bad":"mixed"}function jd({v:o}){const c=Kv(o);if(o.up_ratio==null)return a.jsx("span",{className:"vote-share none",children:"No votes"});const u=Math.round(o.up_ratio*100);return a.jsxs("span",{className:"vote-share "+c,title:`${o.up} up, ${o.down} down of ${o.rated} rated`+(c==="few"?` (fewer than ${Ll} votes: too few to judge)`:""),children:[a.jsx("span",{className:"vote-track","aria-hidden":!0,children:a.jsx("i",{style:{width:u+"%"}})}),a.jsxs("b",{children:[u,"%"]}),c==="few"&&a.jsx("small",{children:"few votes"})]})}function Ad({v:o}){return a.jsxs("span",{className:"vote-counts",children:[a.jsxs("span",{title:"Thumbs up",children:[a.jsx(Hd,{size:13,"aria-label":"up"})," ",Xs(o.up)]}),a.jsxs("span",{title:"Thumbs down",children:[a.jsx(tv,{size:13,"aria-label":"down"})," ",Xs(o.down)]})]})}const Bm={evidence:["Better elsewhere","Another application or model does clearly better on a module of the same name."],escalation:["One Level up","No better evidence yet: the next Level's model, untested for this module."],cheaper:["Cheaper may do","Rated very highly: the Level below may do as well for less, untested for this module."]};async function Fv(o,c,u,r){const h=await fetch(`/api/v1/module-routes/${encodeURIComponent(o)}/${encodeURIComponent(c)}/${encodeURIComponent(u)}`,{method:r?"PUT":"DELETE",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:r?JSON.stringify(r):void 0});if(!h.ok)throw new Error((await h.json().catch(()=>null))?.error?.message||"HTTP "+h.status)}function Xb({items:o,onDone:c}){const[u,r]=f.useState(""),[h,m]=f.useState("");if(!o.length)return null;const g=async v=>{const b=v.application+"/"+v.module+"/"+v.name;r(b),m("");try{await Fv(v.application,v.module,v.name,{connection:v.to.connection,model:v.to.model,why:"from Feedback: "+v.why}),c()}catch(C){m(String(C.message||C))}finally{r("")}};return a.jsxs("section",{className:"panel","aria-labelledby":"suggest-heading",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{id:"suggest-heading",children:"Suggestions from the votes"}),a.jsx("span",{children:o.length===1?"1 suggestion":o.length+" suggestions"})]}),a.jsx("p",{className:"muted feedback-intro",children:"aiport never moves a module by itself. Apply routes only that module's calls for that name to the suggested model, as a Change you can revert on Changes; then watch the module's share of up-votes here."}),h&&a.jsx("div",{className:"notice error feedback-notice",role:"alert",children:h}),a.jsx("ul",{className:"suggestions",children:o.map(v=>{const b=v.application+"/"+v.module+"/"+v.name;return a.jsxs("li",{className:"suggestion "+v.basis,children:[a.jsxs("div",{className:"suggestion-main",children:[a.jsxs("div",{className:"suggestion-title",children:[a.jsx("b",{children:v.application})," · ",a.jsx("code",{children:v.module})," sending ",a.jsx("code",{children:v.name}),a.jsx("span",{className:"badge "+(v.basis==="cheaper"?"good":v.basis==="evidence"?"info":"warn"),title:Bm[v.basis][1],children:Bm[v.basis][0]})]}),a.jsxs("div",{className:"suggestion-move",children:[v.from.model," ",a.jsxs("small",{className:"muted",children:["on ",v.from.connection]})," ",a.jsx(Ld,{size:14,"aria-hidden":!0})," ",a.jsx("b",{children:v.to.model})," ",a.jsxs("small",{className:"muted",children:["on ",v.to.connection]})]}),a.jsxs("p",{className:"muted-small",children:[v.why,"."]})]}),a.jsx("button",{type:"button",className:"primary",disabled:u===b,onClick:()=>g(v),children:u===b?"Applying…":"Apply"})]},b)})})]})}function Vb({app:o,routes:c,onChange:u}){const r=o.modules.filter(g=>g.module&&g.rated>=Ll&&g.up_ratio!=null),h=r.length>1?r.reduce((g,v)=>v.up_ratio>g.up_ratio?v:g):null,m=r.length>1?r.reduce((g,v)=>v.up_ratio<g.up_ratio?v:g):null;return a.jsxs("section",{className:"panel feedback-app","aria-label":"Votes of "+o.application,children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:o.application}),a.jsxs("span",{className:"feedback-summary",children:[a.jsx(Ad,{v:o}),a.jsx(jd,{v:o}),a.jsxs("small",{children:[Xs(o.rated)," of ",Xs(o.calls)," calls rated"]})]})]}),h&&m&&h!==m&&a.jsxs("p",{className:"feedback-callout",children:["Best module: ",a.jsx("b",{children:h.module})," (",Math.round(h.up_ratio*100),"% up). Weakest: ",a.jsx("b",{children:m.module})," (",Math.round(m.up_ratio*100),"% up): try another model or effort for it under Routing."]}),a.jsxs("table",{className:"feedback-table",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Module"}),a.jsx("th",{children:"Sent as"}),a.jsx("th",{children:"Model"}),a.jsx("th",{children:"Effort"}),a.jsx("th",{className:"n",children:"Calls"}),a.jsx("th",{className:"n",children:"Votes"}),a.jsx("th",{children:"Up"})]})}),a.jsx("tbody",{children:o.modules.map(g=>a.jsxs($i.Fragment,{children:[a.jsxs("tr",{className:"section-row "+Kv(g),children:[a.jsxs("th",{scope:"row",colSpan:4,children:[g.module?a.jsx("code",{children:g.module}):a.jsx("span",{className:"muted",children:"No module named"}),g.rows.length>1&&a.jsxs("small",{className:"muted",children:[" · ",g.rows.length," combinations"]})]}),a.jsx("td",{className:"n",children:Xs(g.calls)}),a.jsx("td",{className:"n",children:a.jsx(Ad,{v:g})}),a.jsx("td",{children:a.jsx(jd,{v:g})})]}),g.rows.map(v=>a.jsxs("tr",{className:"combo-row",children:[a.jsx("td",{}),a.jsx("td",{children:a.jsx("code",{children:v.name})}),a.jsx("td",{children:v.model}),a.jsx("td",{children:v.effort||"—"}),a.jsx("td",{className:"n",children:Xs(v.calls)}),a.jsx("td",{className:"n",children:a.jsx(Ad,{v})}),a.jsx("td",{children:a.jsx(jd,{v})})]},v.name+"|"+v.model+"|"+v.effort))]},g.module||"(none)"))})]}),c.length>0&&a.jsxs("div",{className:"module-routes",children:[a.jsx("h3",{children:"Module routes"}),a.jsx("ul",{children:c.map(g=>a.jsxs("li",{children:[a.jsx("code",{children:g.module})," sending ",a.jsx("code",{children:g.name})," ",a.jsx(Ld,{size:13,"aria-hidden":!0})," ",a.jsx("b",{children:g.model})," ",a.jsxs("small",{className:"muted",children:["on ",g.target]}),a.jsxs("button",{type:"button",className:"text-button",onClick:()=>Fv(o.application,g.module,g.name,null).then(u,()=>u()),children:[a.jsx(Id,{size:13,"aria-hidden":!0})," Remove"]})]},g.module+"/"+g.name))})]}),o.modules.some(g=>!g.module)&&a.jsxs("p",{className:"caption",children:["Calls without a module did not send ",a.jsx("code",{children:"X-Aiport-Module"}),". Guidance version 4 asks the application's coding agent to add it."]})]})}function Qb({tick:o,onChange:c}){const[u,r]=f.useState(720),[h,m]=f.useState(null),[g,v]=f.useState(""),[b,C]=f.useState(0),x=()=>{C(A=>A+1),c?.()};f.useEffect(()=>{fetch("/api/v1/feedback?hours="+u,{headers:{"X-Aiport-CSRF":"1"}}).then(async A=>{if(!A.ok)throw new Error((await A.json().catch(()=>null))?.error?.message||"HTTP "+A.status);return A.json()}).then(A=>{m(A),v("")}).catch(A=>v(String(A.message||A)))},[u,o,b]);const y=h?.applications||[],w=y.some(A=>A.rated>0);return a.jsxs(a.Fragment,{children:[a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"How people rate the answers"}),a.jsx("div",{className:"segmented",role:"group","aria-label":"Period",children:Yb.map(([A,E])=>a.jsx("button",{type:"button",className:A===u?"on":"","aria-pressed":A===u,onClick:()=>r(A),children:E},A))})]}),a.jsxs("p",{className:"muted feedback-intro",children:["Each application names the module of every call (",a.jsx("code",{children:h?.module_header||"X-Aiport-Module"}),", such as tldr, classify or insights) and sends its users' thumbs up and down to ",a.jsx("code",{children:h?.feedback_path||"/v1/aiport/feedback"}),". Read each module's share of up-votes by the name it sent, the model that served it and the effort; when a module does badly, change its Level, task name or effort under Routing. Modules with fewer than ",Ll," votes are too few to judge."]}),g&&a.jsx("div",{className:"notice error feedback-notice",role:"alert",children:g}),h&&h.telemetry.state==="disabled"&&a.jsxs("div",{className:"notice",children:[a.jsx("span",{className:"dot amber"}),"Request history is off, so no calls or votes are kept. Turn it on in Settings."]}),h&&!y.length&&a.jsx("div",{className:"quiet-empty feedback-empty",children:"No model calls in this period."}),h&&y.length>0&&!w&&a.jsxs("div",{className:"quiet-empty feedback-empty",children:["No votes yet.",a.jsx("br",{}),a.jsx("small",{children:"Applications start sending them once their coding agent follows guidance version 4 (Integrate → Guidance)."})]})]}),a.jsx(Xb,{items:h?.suggestions||[],onDone:x}),y.map(A=>a.jsx(Vb,{app:A,routes:h?.module_routes?.[A.application]||[],onChange:x},A.application))]})}function Zb({anchor:o=""}){const c=f.useMemo(()=>ew(_d,2),[]),[u,r]=f.useState(""),[h,m]=f.useState("");f.useEffect(()=>{const v=o?document.getElementById(Vs(o)):null;v?v.scrollIntoView({block:"start"}):window.scrollTo(0,0)},[o]),f.useEffect(()=>{const v=c.map(([,C])=>Vs(C)),b=()=>{let C=v[0]||"";for(const x of v){const y=document.getElementById(x);y&&y.getBoundingClientRect().top<window.innerHeight/4&&(C=x)}m(C)};return b(),window.addEventListener("scroll",b,{passive:!0}),()=>window.removeEventListener("scroll",b)},[c]);const g=u.trim()?c.filter(([,v])=>v.toLowerCase().includes(u.trim().toLowerCase())):c;return a.jsxs("div",{className:"docs",children:[a.jsxs("nav",{className:"docs-toc","aria-label":"Manual contents",children:[a.jsxs("label",{className:"docs-filter",children:[a.jsx(Hx,{size:14,strokeWidth:1.75,"aria-hidden":!0}),a.jsx("input",{type:"search",value:u,onChange:v=>r(v.target.value),placeholder:"Filter sections","aria-label":"Filter sections"})]}),a.jsxs("ol",{children:[g.map(([v,b])=>{const C=Vs(b);return a.jsx("li",{className:"level-"+v,children:a.jsx("a",{href:"#"+C,className:h===C?"current":"","aria-current":h===C?"location":void 0,onClick:x=>{x.preventDefault(),document.getElementById(C)?.scrollIntoView({block:"start"})},children:b.replace(/[`*]/g,"")})},C)}),!g.length&&a.jsx("li",{className:"docs-none",children:"No section matches."})]})]}),a.jsx("article",{className:"docs-body brief","aria-label":"aiport manual",children:a.jsx(cv,{source:_d,anchors:!0})})]})}const Ym="aiport.appearance",Xm="aiport.sidebar.collapsed";function Jv(o,c,u){try{const r=localStorage.getItem(o);return c.includes(r)?r:u}catch{return u}}function Wv(o,c){try{localStorage.setItem(o,c)}catch{}}function Kb(){const[o,c]=f.useState(()=>Jv(Ym,["light","dark","system"],"system"));return f.useEffect(()=>{const u=window.matchMedia("(prefers-color-scheme: light)"),r=()=>{document.documentElement.dataset.theme=o==="system"?u.matches?"light":"dark":o};if(r(),Wv(Ym,o),o==="system")return u.addEventListener("change",r),()=>u.removeEventListener("change",r)},[o]),[o,c]}const Fb=[["light","Light",Xx],["dark","Dark",Rx],["system","System",zx]];function Jb({appearance:o,onAppearance:c}){return a.jsx("div",{className:"theme-switch",role:"radiogroup","aria-label":"Appearance",children:Fb.map(([u,r,h])=>a.jsx("button",{type:"button",role:"radio","aria-checked":o===u,"aria-label":r,title:r+" appearance",className:o===u?"on":"",onClick:()=>c(u),children:a.jsx(h,{size:15,strokeWidth:1.75,"aria-hidden":!0})},u))})}const Wb=[["Overview",_x],["Activity",Jm],["Changes",Ig]],Pb=[["Integrate",Dx],["Providers",vx],["Routing",Bg],["Agents",Ud],["Compression",Jg],["Feedback",Hd]],$b=[["Tasks",Ng],["Settings",Gx]];function ej({logo:o,page:c,onPage:u,listen:r,authDisabled:h,onSignOut:m,pending:g=0}){const[v,b]=f.useState(()=>Jv(Xm,["1","0"],"0")==="1");f.useEffect(()=>{Wv(Xm,v?"1":"0"),document.documentElement.dataset.sidebar=v?"collapsed":"expanded"},[v]);const C=([x,y])=>{const w=x==="Changes"?g:0,A=w?x+", "+w+" waiting for a decision":x;return a.jsxs("button",{className:c===x?"active":"","aria-current":c===x?"page":void 0,title:v?A:void 0,"aria-label":v||w?A:void 0,onClick:()=>u(x),children:[a.jsx(y,{size:18,strokeWidth:1.75,"aria-hidden":!0}),a.jsx("span",{className:"label",children:x}),!!w&&a.jsx("span",{className:"nav-badge","aria-hidden":!0,children:w})]},x)};return a.jsxs("aside",{className:"sidebar"+(v?" collapsed":""),children:[a.jsxs("div",{className:"sidebar-top",children:[a.jsxs("div",{className:"brand",children:[a.jsx("img",{className:"service-logo",src:o,alt:"",width:"88",height:"88"}),a.jsx("span",{className:"label",children:"aiport"})]}),a.jsx("button",{className:"icon-button collapse-toggle",onClick:()=>b(x=>!x),"aria-expanded":!v,"aria-label":v?"Expand sidebar":"Collapse sidebar",title:v?"Expand sidebar":"Collapse sidebar",children:v?a.jsx(qx,{size:18,strokeWidth:1.75,"aria-hidden":!0}):a.jsx(Mx,{size:18,strokeWidth:1.75,"aria-hidden":!0})})]}),a.jsxs("nav",{"aria-label":"Main navigation",children:[a.jsx("span",{className:"nav-group label",children:"Monitor"}),Wb.map(C),a.jsx("span",{className:"nav-group label",children:"Configure"}),Pb.map(C)]}),a.jsxs("div",{className:"sidebar-bottom",children:[a.jsx("nav",{"aria-label":"Workspace",children:$b.map(C)}),a.jsxs("div",{className:"sidebar-foot",children:[a.jsxs("div",{className:"local-label",title:r,children:[a.jsx("span",{className:"dot"}),a.jsx("span",{className:"label",children:r||"Connecting…"})]}),h?a.jsxs("button",{className:"icon-text",onClick:()=>u("Settings"),title:"Authentication is off","aria-label":"Authentication is off. Open Settings",children:[a.jsx(Bx,{size:16,strokeWidth:1.75,"aria-hidden":!0}),a.jsx("span",{className:"label",children:"Authentication off"})]}):a.jsxs("button",{className:"icon-text",onClick:m,title:"Sign out","aria-label":"Sign out",children:[a.jsx(Ox,{size:16,strokeWidth:1.75,"aria-hidden":!0}),a.jsx("span",{className:"label",children:"Sign out"})]})]})]})]})}const tj=o=>o.page?o.page+(o.section?" → "+o.section:""):"";function Pv({reason:o,onPage:c}){return o.page?a.jsxs("button",{type:"button",className:"text-button",onClick:()=>c(o.page),children:["Open ",tj(o)," ",a.jsx(dn,{size:14,"aria-hidden":!0})]}):null}function nj({reason:o,codes:c,onPage:u}){const r=c.filter(([,h])=>h);return a.jsxs("div",{className:"reason",role:"group","aria-label":"Why this request failed",children:[a.jsx("strong",{children:o.summary}),a.jsx("p",{children:o.explanation}),o.fix&&a.jsxs("p",{children:[a.jsx("b",{children:"What to do:"})," ",o.fix," ",a.jsx(Pv,{reason:o,onPage:u})]}),a.jsx("small",{children:r.map(([h,m])=>a.jsxs("span",{children:[h," ",a.jsx("code",{children:m})]},h))})]})}function aj({rows:o,onPage:c}){const u=new Map;for(const m of o){const g=u.get(m.route);(!g||m.started_at_ms>g.started_at_ms)&&u.set(m.route,m)}const r=[...u.values()].filter(m=>m.reason?.config).sort((m,g)=>g.started_at_ms-m.started_at_ms)[0];if(!r?.reason)return null;const h=r.reason;return a.jsxs("div",{className:"notice",role:"status",children:[a.jsx("span",{className:"dot amber"}),a.jsxs("div",{children:[a.jsxs("strong",{children:["Recent requests fail. ",h.summary,"."]})," ",h.fix," ",a.jsx(Pv,{reason:h,onPage:c})]})]})}const Vm={header:"from the X-Aiport-Effort header",field:"from the provider's effort field",default:"none stated, so the default"},qd=(o,c="")=>o==null?"—":new Intl.NumberFormat(void 0,{maximumFractionDigits:0}).format(o)+c,sj=o=>o.outcome!=="success"||o.http_status>=400,Qm=2e3;function ij({calls:o,selected:c,onSelect:u,onPage:r,applications:h,recording:m,onSetup:g}){return o.length?a.jsx("div",{className:"table-scroll",children:a.jsxs("table",{className:"activity-table",children:[a.jsx("thead",{children:a.jsxs("tr",{children:[a.jsx("th",{children:"Time"}),a.jsx("th",{children:"Origin / target"}),a.jsx("th",{children:"Model"}),a.jsx("th",{children:"Outcome"}),a.jsx("th",{className:"numeric",children:"Latency"}),a.jsx("th",{})]})}),a.jsx("tbody",{children:o.map(v=>{const b=c===v.id;return a.jsxs($i.Fragment,{children:[a.jsxs("tr",{className:b?"row-open":void 0,children:[a.jsxs("td",{title:new Date(v.started_at_ms).toISOString(),children:[new Date(v.started_at_ms).toLocaleTimeString(),a.jsx("small",{children:new Date(v.started_at_ms).toLocaleDateString()})]}),a.jsxs("td",{children:[v.origin_id||"Unknown",a.jsx("small",{children:v.target_id||"No target"})]}),a.jsxs("td",{children:[v.requested_model||"Unknown",a.jsxs("small",{children:[v.route.replace("/v1/",""),v.resolved_model&&v.resolved_model!==v.requested_model?" → "+v.resolved_model:"",v.effort?" · "+v.effort:"",v.module?" · "+v.module:""]})]}),a.jsxs("td",{children:[a.jsx("span",{className:"badge "+(v.outcome==="success"?"good":""),children:v.outcome.replaceAll("_"," ")}),a.jsx("small",{className:v.reason?"reason-line":void 0,title:v.reason?.explanation,children:v.reason?v.http_status+" · "+v.reason.summary:"HTTP "+v.http_status})]}),a.jsx("td",{className:"numeric",children:qd(v.latency_ms," ms")}),a.jsx("td",{children:a.jsxs("button",{className:"text-button details-toggle"+(b?" on":""),onClick:()=>u(b?"":v.id),"aria-expanded":b,"aria-controls":"detail-"+v.id,"aria-label":(b?"Hide":"Show")+" details of request "+v.id,children:["Details ",a.jsx(xx,{size:14,"aria-hidden":!0})]})})]}),b&&a.jsx("tr",{className:"detail-row",id:"detail-"+v.id,children:a.jsx("td",{colSpan:6,children:a.jsx(oj,{call:v,onPage:r,application:h[v.origin_id]})})})]},v.id)})})]})}):a.jsxs("div",{className:"empty",children:[a.jsx("span",{className:"empty-icon",children:a.jsx(Tx,{size:20,strokeWidth:1.75,"aria-hidden":!0})}),a.jsx("h3",{children:m==="off"?"Activity recording is off":"Your gateway is ready for traffic"}),a.jsx("p",{children:m==="off"?"Enable telemetry in your YAML configuration and restart to see session metadata.":"Connect a coding agent or send an API request. Metadata will appear here; prompts and responses are never captured."}),a.jsx("button",{className:"secondary",onClick:g,children:"View connection setup"})]})}function zt({label:o,children:c,wide:u}){return a.jsxs("div",{className:"detail-field"+(u?" wide":""),children:[a.jsx("dt",{children:o}),a.jsx("dd",{children:c})]})}const va=({v:o})=>o==null||o===""?a.jsx("span",{className:"muted",children:"—"}):a.jsx("code",{children:o});function oj({call:o,onPage:c,application:u}){const r=o.feedback==null?null:o.feedback>.5?"up":o.feedback<.5?"down":"neutral",h=new Date(o.started_at_ms);return a.jsxs("div",{className:"request-detail",role:"region","aria-label":"Details of request "+o.id,children:[o.reason&&a.jsx(nj,{reason:o.reason,codes:[["aiport code",o.refusal_code],["connection error",o.connection_error],["provider error type",o.upstream_error_type],["HTTP status",String(o.http_status)]],onPage:c}),o.requested_model&&a.jsxs("p",{className:"served",children:["Asked for ",a.jsx("b",{children:o.requested_model}),o.effort&&a.jsxs(a.Fragment,{children:[" at ",a.jsx("b",{children:o.effort})," effort (",Vm[o.effort_source||""]||o.effort_source,")"]}),o.resolved_model?a.jsxs(a.Fragment,{children:["; served by ",a.jsx("b",{children:o.target_id||"an unknown Connection"})," with ",a.jsx("b",{children:o.resolved_model}),"."]}):"; not forwarded."]}),a.jsxs("div",{className:"detail-groups",children:[a.jsxs("section",{children:[a.jsx("h4",{children:"Asked"}),a.jsxs("dl",{children:[a.jsx(zt,{label:"Application",children:u?a.jsxs(a.Fragment,{children:[u," ",a.jsx("small",{className:"muted",children:o.origin_id})]}):a.jsx(va,{v:o.origin_id||"Unknown"})}),a.jsx(zt,{label:"Name sent",children:a.jsx(va,{v:o.requested_model})}),a.jsx(zt,{label:"Module",children:a.jsx(va,{v:o.module})}),a.jsx(zt,{label:"Effort",children:o.effort?a.jsxs(a.Fragment,{children:[o.effort," ",a.jsx("small",{className:"muted",children:Vm[o.effort_source||""]||o.effort_source})]}):a.jsx("span",{className:"muted",children:"—"})})]})]}),a.jsxs("section",{children:[a.jsx("h4",{children:"Served"}),a.jsxs("dl",{children:[a.jsx(zt,{label:"Connection",children:a.jsx(va,{v:o.target_id})}),a.jsx(zt,{label:"Model",children:a.jsx(va,{v:o.resolved_model})}),a.jsx(zt,{label:"Provider",children:a.jsx(va,{v:o.provider})}),a.jsx(zt,{label:"Route",children:a.jsx(va,{v:(o.method?o.method+" ":"")+o.route})})]})]}),a.jsxs("section",{children:[a.jsx("h4",{children:"Result"}),a.jsxs("dl",{children:[a.jsxs(zt,{label:"Outcome",children:[a.jsx("span",{className:"badge "+(o.outcome==="success"?"good":"bad"),children:o.outcome.replaceAll("_"," ")})," ",a.jsxs("small",{className:"muted",children:["HTTP ",o.http_status]})]}),a.jsxs(zt,{label:"Latency",children:[qd(o.latency_ms," ms"),o.ttfb_ms!=null&&a.jsxs("small",{className:"muted",children:[" · first byte ",qd(o.ttfb_ms," ms")]})]}),a.jsx(zt,{label:"Vote",children:r==="up"?a.jsxs("span",{className:"vote-chip up",children:[a.jsx(Hd,{size:13,"aria-hidden":!0})," up"]}):r==="down"?a.jsxs("span",{className:"vote-chip down",children:[a.jsx(tv,{size:13,"aria-hidden":!0})," down"]}):r?"neutral":a.jsx("span",{className:"muted",children:"not rated"})}),o.compression&&a.jsxs(zt,{label:"Compression",children:[o.compression," ",a.jsx("small",{className:"muted",children:o.compression_status})]})]})]}),a.jsxs("section",{children:[a.jsx("h4",{children:"Request"}),a.jsxs("dl",{children:[a.jsx(zt,{label:"Id",children:a.jsx(va,{v:o.id})}),a.jsxs(zt,{label:"Time",children:[h.toLocaleString()," ",a.jsx("small",{className:"muted",children:h.toISOString()})]})]})]})]}),sj(o)&&a.jsx(lj,{call:o,application:u,onPage:c}),a.jsx("p",{className:"caption",children:"Prompts, tool arguments and responses are never recorded: only this metadata."})]})}function lj({call:o,application:c,onPage:u}){const[r,h]=f.useState(""),[m,g]=f.useState(!1),[v,b]=f.useState(""),[C,x]=f.useState(null),[y,w]=f.useState(!1),A=c||o.origin_id;if(!c)return a.jsx("div",{className:"report-agent muted-small",children:"This request came from no Application, so there is no coding agent to report it to."});const E=async()=>{g(!0),b("");try{const D=await fetch("/api/v1/issues",{method:"POST",credentials:"same-origin",headers:{"Content-Type":"application/json","X-Aiport-CSRF":"1"},body:JSON.stringify({request_id:o.id,note:r})}),H=await D.json().catch(()=>null);if(!D.ok)throw new Error(H?.error?.message||"HTTP "+D.status);x(H.issue)}catch(D){b(String(D.message||D))}finally{g(!1)}},_=()=>{C&&(history.replaceState(null,"","#"+Xd(C.id)),u("Changes"))},N=async()=>{if(C)try{await navigator.clipboard.writeText(C.application_fix),w(!0),setTimeout(()=>w(!1),1600)}catch{b("The browser did not allow copying.")}};return a.jsxs("section",{className:"report-agent","aria-label":"Report to "+A+"'s coding agent",children:[a.jsxs("h4",{children:["Hand the fix to ",A,"’s coding agent"]}),C?a.jsxs("div",{className:"report-done",role:"status",children:[a.jsxs("p",{children:[a.jsxs("b",{children:["Reported as Issue ",a.jsx("code",{children:C.id}),C.evidence.count>1?" ("+C.evidence.count+" requests of this kind)":"","."]})," ",A,"’s coding agent reads it at the start of its next session (",a.jsx("code",{children:"aiport_my_issues"}),"), and its next failing call of this kind names it in ",a.jsx("code",{children:"X-Aiport-Hint"}),". To have aiport run a coding agent on it now, use ",a.jsx("b",{children:"Fix it"})," on the Issue."]}),a.jsxs("div",{className:"report-actions",children:[a.jsxs("button",{type:"button",className:"secondary",onClick:_,children:["Open the Issue in Changes ",a.jsx(dn,{size:14,"aria-hidden":!0})]}),a.jsxs("button",{type:"button",className:"text-button",onClick:N,children:[a.jsx(Fa,{size:13,"aria-hidden":!0})," ",y?"Copied":"Copy the fix for a coding agent"]})]})]}):a.jsxs(a.Fragment,{children:[a.jsxs("p",{className:"muted-small",children:["aiport writes what failed, what it knows about it and what to do as an Issue for ",A,". It keeps no provider messages, so paste the error the application showed if you have it."]}),a.jsxs("label",{className:"report-note-field",children:[a.jsxs("span",{children:["Note for the coding agent ",a.jsx("small",{className:"muted",children:"(optional)"})]}),a.jsx("textarea",{value:r,maxLength:Qm,rows:3,placeholder:"For example: "+(o.http_status===400?"Meta answered 400: … Missing 'visualLayout'.":"what the user did and the error shown"),onChange:D=>h(D.target.value)}),a.jsxs("small",{className:"muted",children:[r.length," / ",Qm]})]}),v&&a.jsx("div",{className:"notice error",role:"alert",children:v}),a.jsxs("button",{type:"button",className:"primary report-send",disabled:m,onClick:E,children:[a.jsx("span",{children:m?"Reporting…":"Report to "+A+"’s coding agent"}),a.jsx(Vg,{size:15,"aria-hidden":!0})]})]})]})}const kd=3e3,rj=2*6e4,ya=18,cj=12,dj=900,uj=160,hj=900,pj=700,fj=1400,Kd=(o,c,u)=>Math.max(c,Math.min(u,o)),Ki=(o,c)=>1-Math.exp(-Math.max(0,o)/c),El=o=>o.outcome!=="success"||o.http_status>=400;function mj(o,c){const u=o.at==null?16:Kd(c-o.at,0,250);for(o.at=c;o.queue.length&&o.queue[0].t<=c;){const r=o.queue.shift();o.hit=Math.max(o.hit,r.v),r.failed&&(o.fail=1)}return o.level=o.level<o.hit?o.level+(o.hit-o.level)*Ki(u,cj):o.level*(1-Ki(u,dj)),o.hit*=1-Ki(u,uj),o.level>=o.peak?(o.peak=o.level,o.peakAt=c):c-o.peakAt>hj&&(o.peak*=1-Ki(u,pj)),o.fail*=1-Ki(u,fj),o.level<.003&&(o.level=0),o.peak<.004&&(o.peak=0),o.fail<.01&&(o.fail=0),o.queue.length>0||o.level>0||o.peak>0||o.fail>0}function Ol(o,c,u){return getComputedStyle(o).getPropertyValue(c).trim()||u}function Zm(o,c){const u=o.getContext("2d");if(!u)return;const r=window.devicePixelRatio||1,h=o.clientWidth,m=o.clientHeight;(o.width!==Math.round(h*r)||o.height!==Math.round(m*r))&&(o.width=Math.round(h*r),o.height=Math.round(m*r)),u.setTransform(r,0,0,r,0,0),u.clearRect(0,0,h,m);const g=Ol(o,"--good","#a3be8c"),v=Ol(o,"--warn","#ebcb8b"),b=Ol(o,"--bad","#bf616a"),C=Ol(o,"--line","#3b4252"),x=2,y=(h-x*(ya-1))/ya,w=c.level*ya,A=Math.round(c.peak*ya)-1;for(let E=0;E<ya;E++){const _=E/(ya-1),N=_>.82?b:_>.6?v:g,D=Kd(w-E,0,1),H=E*(y+x);u.globalAlpha=1,u.fillStyle=C,u.beginPath(),u.roundRect(H,1,y,m-2,1.5),u.fill(),(D>0||E===A)&&(u.globalAlpha=E===A&&D===0?.75:.25+.75*D,u.fillStyle=N,u.beginPath(),u.roundRect(H,1,y,m-2,1.5),u.fill())}if(c.fail>0){u.globalAlpha=.8*c.fail,u.fillStyle=b;const E=Math.floor(ya*.83);for(let _=E;_<ya;_++)u.beginPath(),u.roundRect(_*(y+x),1,y,m-2,1.5),u.fill()}u.globalAlpha=1}function gj({apps:o,listen:c}){const[u,r]=f.useState(null),[h,m]=f.useState(!1),[g,v]=f.useState([]),[b,C]=f.useState(Date.now()),x=f.useRef(null),y=f.useRef({level:0,hit:0,peak:0,peakAt:0,fail:0,queue:[],at:null}),w=f.useRef(null),A=f.useRef(0),E=()=>{if(A.current)return;const Y=S=>{const U=x.current,Z=mj(y.current,S);U&&Zm(U,y.current),A.current=Z?requestAnimationFrame(Y):0};A.current=requestAnimationFrame(Y)};f.useEffect(()=>{let Y=!1,S=0;const U=async()=>{if(!Y){if(!document.hidden)try{const be=await fetch("/api/v1/requests?hours=1&limit=100",{credentials:"same-origin"});if(be.status===401){m(!0),r(!0),S=window.setTimeout(U,kd);return}if(!be.ok)throw new Error(String(be.status));m(!1);const J=(await be.json()).requests||[];if(r(!0),v(J),C(Date.now()),w.current){const ve=J.filter(W=>!w.current.has(W.id)).sort((W,F)=>W.started_at_ms-F.started_at_ms);if(ve.length){const W=ve[0].started_at_ms,F=Math.max(1,ve[ve.length-1].started_at_ms-W),ne=performance.now(),le=Kd(.45+.12*Math.log2(1+ve.length),.45,1);ve.forEach(je=>y.current.queue.push({t:ne+(je.started_at_ms-W)/F*Math.min(F,kd-300),v:El(je)?1:le,failed:El(je)})),y.current.queue.sort((je,Ae)=>je.t-Ae.t),E()}}w.current=new Set(J.map(ve=>ve.id))}catch{r(!1)}S=window.setTimeout(U,kd)}};U();const Z=()=>{x.current&&Zm(x.current,y.current)},$=matchMedia("(prefers-color-scheme: dark)");$.addEventListener("change",Z);const oe=new MutationObserver(Z);return oe.observe(document.documentElement,{attributes:!0,attributeFilter:["data-theme","class"]}),Z(),()=>{Y=!0,clearTimeout(S),cancelAnimationFrame(A.current),A.current=0,$.removeEventListener("change",Z),oe.disconnect()}},[]);const _=new Map;for(const Y of g){const S=_.get(Y.origin_id);(!S||Y.started_at_ms>S.started_at_ms)&&_.set(Y.origin_id,Y)}const N=g.filter(Y=>b-Y.started_at_ms<6e4),D=N.filter(El).length,H=o.map(Y=>{const S=_.get(Y.id),U=!!S&&b-S.started_at_ms<rj;return{...Y,last:S,state:S?U?El(S)?"failing":"live":"quiet":"idle"}}),G=H.filter(Y=>Y.state==="live"||Y.state==="failing").length,Q=Y=>{const S=Math.round((b-Y)/1e3);return S<60?S+" s ago":Math.round(S/60)+" min ago"};return a.jsxs("div",{className:"live-gauge",role:"group","aria-label":"Gateway listener, connected applications and live traffic",children:[a.jsxs("span",{className:"gauge-pill"+(u?" on":u===!1?" bad":""),role:"status",title:u===!1?"The dashboard got no answer from the gateway.":"aiport is listening"+(c?" on "+c:""),children:[a.jsx("span",{className:"led","aria-hidden":!0}),u===!1?"No reply":h?"Signed out":u?"Listening":"Connecting…"]}),a.jsxs("span",{className:"gauge-apps",title:H.length?H.map(Y=>Y.name+": "+(Y.last?(Y.state==="failing"?"last call failed, ":"")+Q(Y.last.started_at_ms):"no call in the last hour")).join(`
`):"No Applications configured yet","aria-label":G+" of "+H.length+" applications called in the last two minutes",children:[H.slice(0,8).map(Y=>a.jsx("span",{className:"app-led "+Y.state,"aria-hidden":!0},Y.id)),a.jsx("small",{children:H.length?G+"/"+H.length+" apps":"no apps"})]}),a.jsx("canvas",{ref:x,className:"gauge-meter",role:"img","aria-label":N.length+" calls in the last minute"+(D?", "+D+" failed":""),title:N.length+" calls in the last minute"+(D?", "+D+" failed":"")+". The lights follow each call as it passes; red at the top is a failure."})]})}const vj={openai:"OpenAI",anthropic:"Anthropic",bedrock:"Amazon Bedrock",meta:"Meta Model API",ollama:"Ollama",llamacpp:"llama.cpp"},Gs=(o,c="")=>o==null?"—":new Intl.NumberFormat(void 0,{maximumFractionDigits:0}).format(o)+c,yj=["Overview","Activity","Changes","Integrate","Providers","Routing","Agents","Compression","Feedback","Tasks","Settings","Manual"],Km=()=>yj.find(o=>o.toLowerCase()===location.pathname.replace(/^\/+|\/+$/g,"").toLowerCase())||"Overview";function xj(){const[o,c]=f.useState(""),[u,r]=f.useState(Km),[h,m]=f.useState(""),g=K=>{r(K);const me="/"+K.toLowerCase();location.pathname!==me&&history.pushState(null,"",me+(K==="Changes"?location.hash:""))};f.useEffect(()=>{const K=()=>r(Km());return addEventListener("popstate",K),()=>removeEventListener("popstate",K)},[]);const v=K=>{m(K),g("Manual")},[b,C]=Kb(),[x,y]=f.useState(null),[w,A]=f.useState({}),[E,_]=f.useState({}),[N,D]=f.useState([]),[H,G]=f.useState([]),[Q,Y]=f.useState(null),[S,U]=f.useState(null),[Z,$]=f.useState(""),[oe,be]=f.useState(!1),[re,J]=f.useState("24"),[ve,W]=f.useState(""),[F,ne]=f.useState(""),[le,je]=f.useState(""),[Ae,Qe]=f.useState(""),[Ze,dt]=f.useState(""),[j,I]=f.useState(""),[ee,ie]=f.useState(""),[fe,ye]=f.useState(0),xe=iw(!!o,fe),[ae,he]=f.useState(null);f.useEffect(()=>{o&&de("decisions").then(K=>he((K.items||[]).map(me=>me.id))).catch(()=>he([]))},[o,fe]);const[ut,Qt]=f.useState(0);f.useEffect(()=>{o&&de("runs").then(K=>Qt((K.runs||[]).reduce((me,Be)=>me+(Be.waiting||0),0))).catch(()=>Qt(0))},[o,fe]);const B=(ae?.length||0)+ut,Re=Sv(fe);rw(ae);const Ie=f.useRef(null),we=f.useRef({from:Date.now()-24*36e5,to:Date.now()}),st=()=>new URLSearchParams({hours:re,origin:ve,target:F,outcome:le,model:Ae,from_ms:String(we.current.from),to_ms:String(we.current.to)});async function de(K,me){const Be=await fetch("/api/v1/"+K,{credentials:"same-origin",signal:me});if(Be.status===401)throw new Error("Your session has expired. Sign in again or use token recovery.");if(!Be.ok)throw new Error("The gateway could not complete this request ("+Be.status+").");return Be.json()}f.useEffect(()=>{if(!o)return;const K=Date.now();we.current={from:K-Number(re)*36e5,to:K};const me=new AbortController;return Ie.current?.abort(),Ie.current=me,be(!0),$(""),Promise.all([de("config",me.signal),de("stats?"+st(),me.signal),de("requests?"+st(),me.signal)]).then(([Be,ja,Bn])=>{me.signal.aborted||(y(Be.config),A(Be.secret_status),_(Be.secret_source||{}),D(Be.providers||[]),Y(ja.summary),U(ja.telemetry),G(Bn.requests),ie(Bn.next_cursor))}).catch(Be=>{Be.name!=="AbortError"&&$(Be.message)}).finally(()=>{me.signal.aborted||be(!1)}),()=>me.abort()},[o,re,ve,F,le,Ae,fe]),f.useEffect(()=>{if(!o)return;const K=setInterval(()=>{document.hidden||ye(me=>me+1)},1e4);return()=>clearInterval(K)},[o]);async function Fe(){if(ee){be(!0);try{const K=st();K.set("cursor",ee);const me=await de("requests?"+K);G(Be=>[...Be,...me.requests.filter(ja=>!Be.some(Bn=>Bn.id===ja.id))]),ie(me.next_cursor)}catch(K){$(K.message)}finally{be(!1)}}}async function Ve(){try{await Ji("logout",{})}catch{}Ie.current?.abort(),c(""),y(null),G([]),Y(null),U(null),I(""),$("")}const nn=[...new Set(H.map(K=>K.requested_model).filter(K=>!!K))].sort(),Gn=a.jsxs("div",{className:"filters",children:[a.jsxs("label",{children:["Period",a.jsxs("select",{value:re,onChange:K=>J(K.target.value),children:[a.jsx("option",{value:"1",children:"Last hour"}),a.jsx("option",{value:"24",children:"Last 24 hours"}),a.jsx("option",{value:"168",children:"Last 7 days"})]})]}),a.jsxs("label",{children:["Origin",a.jsxs("select",{value:ve,onChange:K=>W(K.target.value),children:[a.jsx("option",{value:"",children:"All origins"}),a.jsx("option",{value:"unknown",children:"Unknown"}),x?.origins.map(K=>a.jsx("option",{value:K.id,children:K.name},K.id))]})]}),a.jsxs("label",{children:["Target",a.jsxs("select",{value:F,onChange:K=>ne(K.target.value),children:[a.jsx("option",{value:"",children:"All targets"}),x?.targets.map(K=>a.jsx("option",{children:K.id},K.id))]})]}),a.jsxs("label",{children:["Outcome",a.jsxs("select",{value:le,onChange:K=>je(K.target.value),children:[a.jsx("option",{value:"",children:"All outcomes"}),["success","upstream_error","gateway_error","incomplete","timeout","cancelled","rejected"].map(K=>a.jsx("option",{children:K},K))]})]}),a.jsxs("label",{children:["Name",a.jsx("input",{list:"activity-names",placeholder:"Any name",value:Ze,onChange:K=>{const me=K.target.value;dt(me),(!me||nn.includes(me))&&Qe(me.trim())},onBlur:()=>Qe(Ze.trim()),onKeyDown:K=>{K.key==="Enter"&&Qe(Ze.trim())},"aria-label":"Filter by the name sent"}),a.jsx("datalist",{id:"activity-names",children:nn.map(K=>a.jsx("option",{value:K},K))})]})]}),Js=Object.fromEntries((x?.origins||[]).map(K=>[K.id,K.name])),ba=K=>a.jsx(ij,{calls:K,selected:j,onSelect:I,onPage:g,applications:Js,recording:S?.state==="disabled"?"off":"on",onSetup:()=>g("Settings")}),Ws=Date.now(),Ja=Array.from({length:Number(re)},(K,me)=>{const Be=Math.floor(Ws/36e5)*36e5-(Number(re)-1-me)*36e5;return{time:Be,count:Q?.traffic[Be]||0}}),hn=Math.max(1,...Ja.map(K=>K.count));return o?a.jsx(dv.Provider,{value:v,children:a.jsxs("div",{className:"shell",children:[a.jsx(ej,{logo:Cd,page:u,onPage:K=>{g(K),I("")},listen:Im(x),authDisabled:!!x?.auth.disabled,onSignOut:Ve,pending:B}),a.jsxs("main",{className:"main",children:[a.jsxs("header",{children:[a.jsxs("div",{children:[a.jsx("p",{className:"eyebrow",children:"LOCAL AI GATEWAY"}),a.jsxs("h1",{children:[u,Nm[u]&&a.jsx(wt,{section:Nm[u]})]}),a.jsx("p",{className:"muted",children:u==="Overview"?"The shape of your AI traffic, at a glance.":u==="Activity"?"Request metadata. No prompts, no response bodies.":u==="Tasks"?"Capture work for coding agents. You own intent and priority.":u==="Providers"?"The model providers aiport forwards to, on this machine and remote.":u==="Routing"?"Which connection serves each protocol, and how applications reach aiport.":u==="Integrate"?"Connect an application to aiport, step by step, with instructions aiport writes for coding agents.":u==="Manual"?"How to use aiport, page by page.":u==="Agents"?"Coding agents launched through aiport, each with its own token.":u==="Compression"?"Opt-in request compression, and whether it hurts responses.":u==="Feedback"?"Thumbs up and down from the people using each application, per module, model and effort.":u==="Changes"?"Issues and Suggestions waiting for you, every configuration change made through aiport, and a way back.":"Sign-in and the configuration of this aiport install."})]}),a.jsxs("div",{className:"header-side",children:[a.jsx(gj,{apps:(x?.origins||[]).map(K=>({id:K.id,name:K.name})),listen:Im(x)}),a.jsxs("div",{className:"header-actions toolbar",role:"toolbar","aria-label":"Page tools",children:[u==="Tasks"&&a.jsxs(a.Fragment,{children:[a.jsx(ow,{enabled:xe.enabled,onChange:xe.setEnabled}),a.jsx("span",{className:"toolbar-sep","aria-hidden":!0})]}),a.jsx("button",{type:"button",className:"icon-button help-button"+(u==="Manual"?" on":""),onClick:()=>{m(""),g("Manual")},title:"Manual","aria-label":"Open the aiport manual",children:a.jsx(jx,{size:16,strokeWidth:1.75,"aria-hidden":!0})}),a.jsx("span",{className:"toolbar-sep","aria-hidden":!0}),a.jsx(Jb,{appearance:b,onAppearance:C}),a.jsx("span",{className:"toolbar-sep","aria-hidden":!0}),a.jsx("button",{type:"button",className:"icon-button",disabled:oe,onClick:()=>ye(K=>K+1),title:"Refresh now. Pages also update by themselves every 10 seconds.","aria-label":oe?"Refreshing":"Refresh now",children:a.jsx(Ux,{size:16,strokeWidth:1.75,"aria-hidden":!0,className:oe?"spin":""})})]})]})]}),Z&&a.jsxs("div",{className:"notice error",role:"alert",children:[Z,a.jsx("button",{onClick:Ve,children:"Sign in again"})]}),(u==="Overview"||u==="Activity")&&a.jsxs("div",{className:"notice",children:[a.jsx("span",{className:"dot amber"}),a.jsxs("div",{children:[S?.state==="disabled"?a.jsxs(a.Fragment,{children:[a.jsx("strong",{children:"Request recording is off."}),a.jsxs("span",{children:[" Your configuration sets ",a.jsx("code",{children:"telemetry.enabled: false"}),", so aiport forwards requests but records nothing here. Set it to ",a.jsx("code",{children:"true"})," and restart to see activity. Only metadata is recorded, never prompts or responses."]})]}):S?.persisting?a.jsxs(a.Fragment,{children:[a.jsx("strong",{children:"Activity is saved on this machine."}),a.jsxs("span",{children:[" Request metadata is kept in a private SQLite file next to your configuration for ",S.retention_days??7," days. Prompts and responses are never stored. Token and cost figures aren't measured yet."]}),!!S.failed_writes&&a.jsxs("b",{children:[" ",S.failed_writes," records could not be saved."]})]}):a.jsxs(a.Fragment,{children:[a.jsx("strong",{children:"Activity is kept for this session only."}),a.jsx("span",{children:" The history database could not be opened, so the latest 500 calls are held in memory and cleared on restart. Token and cost figures aren't measured yet."})]}),!!((S?.dropped_events||0)+(S?.evicted_events||0))&&a.jsxs("b",{children:[" Incomplete history: ",S?.dropped_events||0," dropped, ",S?.evicted_events||0," evicted."]})]})]}),(u==="Overview"||u==="Activity")&&Gn,u==="Overview"&&B>0&&a.jsxs("div",{className:"notice",role:"status",children:[a.jsx("span",{className:"dot amber"}),a.jsxs("div",{children:[a.jsxs("strong",{children:[B===1?"1 item waits":B+" items wait"," for your decision."]})," ",a.jsxs("button",{className:"text-button",onClick:()=>g("Changes"),children:["Review in Changes ",a.jsx(dn,{size:14,"aria-hidden":!0})]})]})]}),u==="Overview"&&!!Re?.behind&&a.jsxs("div",{className:"notice",role:"status",children:[a.jsx("span",{className:"dot info"}),a.jsxs("div",{children:[Lw(Re)," ",a.jsxs("button",{className:"text-button",onClick:()=>g("Integrate"),children:["See Guidance in Integrate ",a.jsx(dn,{size:14,"aria-hidden":!0})]})]})]}),u==="Overview"&&a.jsx(aj,{rows:H,onPage:g}),u==="Overview"&&a.jsxs(a.Fragment,{children:[a.jsx("section",{className:"metrics","aria-label":"Traffic summary",children:[["Requests",Gs(Q?.requests),"In the selected window"],["Average latency",Gs(Q?.average_latency_ms," ms"),"End-to-end forwarding time"],["P95 latency",Gs(Q?.p95_latency_ms," ms"),Gs(Q?.requests)+" latency samples"],["Error rate",Q?(Q.error_rate*100).toFixed(1)+"%":"—",Gs(Q?.cancellations)+" cancellations · "+Gs(Q?.rejections)+" rejected"]].map(([K,me,Be])=>a.jsxs("article",{className:"metric",children:[a.jsx("span",{children:K}),a.jsx("strong",{children:me}),a.jsx("small",{children:Be})]},K))}),a.jsxs("div",{className:"charts",children:[a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Request traffic"}),a.jsxs("span",{children:["Last ",re," hours · hourly"]})]}),a.jsx("div",{className:"bars",style:{gap:Number(re)>24?2:7},role:"img","aria-label":"Hourly request counts over the selected window, "+Ja.reduce((K,me)=>K+me.count,0)+" retained calls",children:Ja.map(K=>a.jsx("div",{title:new Date(K.time).toLocaleString()+": "+K.count+" requests",children:a.jsx("i",{style:{height:Math.max(K.count?4:1,K.count/hn*100)+"%"}})},K.time))}),a.jsxs("div",{className:"axis",children:[a.jsxs("span",{children:[re," hours ago"]}),a.jsx("span",{children:"Now"})]})]}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Provider distribution"}),a.jsx("span",{children:"Retained calls"})]}),Object.keys(Q?.providers||{}).length?Object.entries(Q.providers).map(([K,me])=>a.jsxs("div",{className:"distribution",children:[a.jsxs("div",{children:[a.jsxs("span",{className:"with-logo",children:[K&&a.jsx(Ks,{provider:K,location:N.find(Be=>Be.id===K)?.location,size:14}),vj[K]||K||"Unknown"]}),a.jsx("strong",{children:me})]}),a.jsx("div",{className:"track",children:a.jsx("i",{style:{width:me/Q.requests*100+"%"}})})]},K)):a.jsxs("div",{className:"quiet-empty",children:["No provider traffic yet.",a.jsx("br",{}),a.jsx("small",{children:"Distribution appears after your first request."})]})]})]}),a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Recent requests"}),a.jsxs("button",{className:"text-button",onClick:()=>g("Activity"),children:["View activity ",a.jsx(dn,{size:14,"aria-hidden":!0})]})]}),ba(H.slice(0,5))]})]}),u==="Tasks"&&a.jsx(Ub,{refresh:fe}),u==="Changes"&&a.jsx(eb,{tick:fe,targets:x?.targets.map(K=>K.id)||[],onChange:()=>ye(K=>K+1),onOpenApplications:()=>g("Integrate")}),u==="Feedback"&&a.jsx(Qb,{tick:fe,onChange:()=>ye(K=>K+1)}),u==="Compression"&&a.jsx(Bb,{tick:fe,config:x,providers:N,onChange:()=>ye(K=>K+1)}),u==="Activity"&&a.jsxs("section",{className:"panel",children:[a.jsxs("div",{className:"panel-heading",children:[a.jsx("h2",{children:"Request activity"}),a.jsxs("span",{children:[H.length," loaded · local timezone"]})]}),ba(H),ee&&a.jsx("div",{className:"load-more",children:a.jsx("button",{className:"secondary",disabled:oe,onClick:Fe,children:"Load older requests"})})]}),(()=>{const K={config:x,providers:N,secrets:w,sources:E,onKeys:me=>{A(me.secret_status),_(me.secret_source)},onChange:()=>ye(me=>me+1),queueDepth:S?.queue_depth};return u==="Providers"?a.jsx(_b,{...K}):u==="Routing"?a.jsx(Eb,{...K,onIntegrate:()=>g("Integrate")}):u==="Integrate"?a.jsx(Xw,{origins:x?.origins||[],targets:x?.targets||[],providers:N,tick:fe,onChange:()=>ye(me=>me+1),onManual:()=>g("Manual"),onChanges:()=>g("Changes")}):u==="Manual"?a.jsx(Zb,{anchor:h}):u==="Agents"?a.jsx(Ob,{...K}):u==="Settings"?a.jsx(Rb,{...K}):null})(),xe.done&&a.jsx(lw,{blocked:xe.done.blocked,onOpen:()=>{g("Tasks"),xe.dismiss()},onClose:xe.dismiss}),a.jsxs("footer",{children:["aiport · Developer preview",a.jsx("span",{children:"Metadata only · Polls every 10 seconds"})]})]})]})}):a.jsx(Fx,{onAuthenticated:()=>c("session")})}sx.createRoot(document.getElementById("root")).render(a.jsx(xj,{}));
