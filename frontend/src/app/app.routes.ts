import { Routes } from '@angular/router';
import { TopicsComponent } from './topics/topics.component';
import { TopicViewComponent } from './topic-view/topic-view.component';

export const routes: Routes = [
    { path: '', redirectTo: 'topics', pathMatch: 'full' },
    { path: 'topics', component: TopicsComponent },
    { path: 'topics/:id', component: TopicViewComponent }
];
